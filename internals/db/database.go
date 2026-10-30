package db

import (
	"chat-server/internals/db/models"
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
var DB *gorm.DB

func ConnectToDB() *gorm.DB {

	dbe := os.Getenv("DATABASE_URL")
	// runMigrations := os.Getenv("RUN_MIGRATIONS")
	if dbe == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// -- ConnectToDB Database ------
	var err error
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN: dbe,
		// PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to Connect to the DATABASE")
	}
	log.Println("DATABASE Connection sucessful.")

	//---Run AutoMigrate ---
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		log.Println("Running migrations...")
		if err := migrateChannels(DB); err != nil {
			log.Fatalf("Channel rename migration failed: %v", err)
		}
		if err := DB.AutoMigrate(
			&models.User{},
			&models.Account{},
			&models.Session{},
			&models.Channel{},
			&models.Message{},
			&models.MessageUserDeletion{},
			&models.Plan{},
			&models.PlanParticipant{},
			&models.PlanPoll{},
			&models.PlanPollOption{},
			&models.PlanPollVote{},
			&models.Friend{},
			&models.ChannelMember{},
		); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		if err := migrateUserCascadeDeletes(DB); err != nil {
			log.Fatalf("User cascade migration failed: %v", err)
		}
		if err := migratePlanStatuses(DB); err != nil {
			log.Fatalf("Plan status migration failed: %v", err)
		}
		return DB
	}
	// runMig := false

	// if runMig {
	// 	if os.Getenv("APP_ENV") == "local" {
	// 		if err := DB.AutoMigrate(
	// 			&models.User{},
	// 			&models.Account{},
	// 			&models.Session{},
	// 			&models.Channel{},
	// 			&models.Message{},
	// 			&models.Friend{},
	// 		); err != nil {
	// 			log.Fatalf("Migration failed: %v", err)
	// 		}
	// 		log.Println("Database migration successful")
	// 	} else {
	// 		log.Println("-----Skiping the migration-------")
	// 	}
	//
	// }

	return DB
}

// migrateChannels renames the legacy crew schema to the channel schema while
// keeping existing rows. It is safe to run more than once.
func migrateChannels(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`
			DO $$
			BEGIN
				IF to_regclass('crews') IS NOT NULL AND to_regclass('channels') IS NULL THEN
					ALTER TABLE crews RENAME TO channels;
				END IF;
				IF to_regclass('crew_members') IS NOT NULL AND to_regclass('channel_members') IS NULL THEN
					ALTER TABLE crew_members RENAME TO channel_members;
				END IF;
				IF EXISTS (
					SELECT 1 FROM information_schema.columns
					WHERE table_schema = current_schema() AND table_name = 'channel_members' AND column_name = 'crew_id'
				) AND NOT EXISTS (
					SELECT 1 FROM information_schema.columns
					WHERE table_schema = current_schema() AND table_name = 'channel_members' AND column_name = 'channel_id'
				) THEN
					ALTER TABLE channel_members RENAME COLUMN crew_id TO channel_id;
				END IF;
				IF EXISTS (
					SELECT 1 FROM information_schema.columns
					WHERE table_schema = current_schema() AND table_name = 'messages' AND column_name = 'crew_id'
				) AND NOT EXISTS (
					SELECT 1 FROM information_schema.columns
					WHERE table_schema = current_schema() AND table_name = 'messages' AND column_name = 'channel_id'
				) THEN
					ALTER TABLE messages RENAME COLUMN crew_id TO channel_id;
				END IF;
				IF EXISTS (
					SELECT 1 FROM information_schema.columns
					WHERE table_schema = current_schema() AND table_name = 'plans' AND column_name = 'crew_id'
				) AND NOT EXISTS (
					SELECT 1 FROM information_schema.columns
					WHERE table_schema = current_schema() AND table_name = 'plans' AND column_name = 'channel_id'
				) THEN
					ALTER TABLE plans RENAME COLUMN crew_id TO channel_id;
				END IF;
			END $$;
		`).Error; err != nil {
			return err
		}

		return renameLegacyChannelObjects(tx)
	})
}

// renameLegacyChannelObjects renames indexes and constraints that still carry
// the old crew name after the tables and columns have been renamed.
func renameLegacyChannelObjects(tx *gorm.DB) error {
	var indexes []string
	if err := tx.Raw(`
		SELECT indexname FROM pg_indexes
		WHERE schemaname = current_schema() AND indexname LIKE '%crew%'
	`).Pluck("indexname", &indexes).Error; err != nil {
		return err
	}
	for _, name := range indexes {
		newName := strings.ReplaceAll(name, "crew", "channel")
		if err := tx.Exec(fmt.Sprintf(
			"ALTER INDEX %s RENAME TO %s", quoteIdentifier(name), quoteIdentifier(newName),
		)).Error; err != nil {
			return err
		}
	}

	var constraints []struct {
		TableName      string
		ConstraintName string
	}
	if err := tx.Raw(`
		SELECT conrelid::regclass::text AS table_name, conname AS constraint_name
		FROM pg_constraint
		WHERE conname LIKE '%crew%'
	`).Scan(&constraints).Error; err != nil {
		return err
	}
	for _, constraint := range constraints {
		newName := strings.ReplaceAll(constraint.ConstraintName, "crew", "channel")
		if err := tx.Exec(fmt.Sprintf(
			"ALTER TABLE %s RENAME CONSTRAINT %s TO %s",
			constraint.TableName, quoteIdentifier(constraint.ConstraintName), quoteIdentifier(newName),
		)).Error; err != nil {
			return err
		}
	}

	return nil
}

// migrateUserCascadeDeletes makes every foreign key that references users
// cascade on delete. Without this, purging an unverified user fails when the
// user still owns a channel, is a channel member, or has messages/sessions.
func migrateUserCascadeDeletes(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(`
			DO $$
			DECLARE
				fk record;
			BEGIN
				FOR fk IN
					SELECT c.conname,
					       c.conrelid::regclass::text AS table_name,
					       regexp_replace(
					           pg_get_constraintdef(c.oid),
					           ' ON DELETE (NO ACTION|RESTRICT)',
					           ''
					       ) AS definition
					FROM pg_constraint c
					WHERE c.contype = 'f'
					  AND c.confrelid = 'users'::regclass
					  AND c.confdeltype <> 'c'
				LOOP
					EXECUTE format(
						'ALTER TABLE %s DROP CONSTRAINT %I',
						fk.table_name, fk.conname
					);
					EXECUTE format(
						'ALTER TABLE %s ADD CONSTRAINT %I %s ON DELETE CASCADE',
						fk.table_name, fk.conname, fk.definition
					);
				END LOOP;
			END $$;
		`).Error
	})
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func migratePlanStatuses(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("ALTER TABLE plans DROP CONSTRAINT IF EXISTS chk_plans_status").Error; err != nil {
			return err
		}
		if err := tx.Exec("UPDATE plans SET status = 'open' WHERE status = 'scheduled'").Error; err != nil {
			return err
		}
		if err := tx.Exec("ALTER TABLE plans ALTER COLUMN status SET DEFAULT 'open'").Error; err != nil {
			return err
		}
		return tx.Exec("ALTER TABLE plans ADD CONSTRAINT chk_plans_status CHECK (status = 'open' OR status = 'confirmed' OR status = 'completed' OR status = 'cancelled')").Error
	})
}
