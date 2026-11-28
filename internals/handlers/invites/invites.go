package invites

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"chat-server/internals/db/models"
	"chat-server/internals/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Handler struct {
	DB      *gorm.DB
	Limits  *Limits
	Publish func(uuid.UUID, ...uuid.UUID)
}

type problem struct {
	code int
	text string
}

func (p problem) Error() string { return p.text }

var missing = problem{http.StatusNotFound, "invite or user not found"}
var denied = problem{http.StatusForbidden, "channel owner or admin required"}
var closed = problem{http.StatusConflict, "invite is no longer pending"}
var expired = problem{http.StatusConflict, "invite has expired"}

type inviteView struct {
	ID          uuid.UUID `json:"id"`
	ChannelID   uuid.UUID `json:"channelId"`
	ChannelName string    `json:"channelName"`
	InviterName string    `json:"inviterName"`
	InviteeName string    `json:"inviteeName"`
	InviteeID   uuid.UUID `json:"inviteeId"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

type targetView struct {
	ID           uuid.UUID `json:"id"`
	PublicID     string    `json:"publicId"`
	Name         string    `json:"name"`
	Relationship string    `json:"relationship"`
}

func (h Handler) Lookup(w http.ResponseWriter, r *http.Request) { h.target(w, r, false) }
func (h Handler) Create(w http.ResponseWriter, r *http.Request) { h.target(w, r, true) }

func (h Handler) target(w http.ResponseWriter, r *http.Request, create bool) {
	actor, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		fail(w, problem{401, "unauthorized"})
		return
	}
	var input struct {
		GroupID  uuid.UUID `json:"groupId"`
		PublicID uuid.UUID `json:"publicId"`
	}
	if err := readJSON(w, r, &input); err != nil || input.PublicID == uuid.Nil {
		fail(w, problem{400, "a valid public ID is required"})
		return
	}
	if create {
		input.GroupID, _ = uuid.Parse(chi.URLParam(r, "channelID"))
	}
	if input.GroupID == uuid.Nil {
		fail(w, problem{400, "invalid channel"})
		return
	}
	var view targetView
	now := time.Now().UTC()
	if h.Limits != nil && !h.Limits.allow(actor, now) {
		w.Header().Set("Retry-After", "2")
		fail(w, problem{http.StatusTooManyRequests, "too many invite requests; try again shortly"})
		return
	}
	err = h.DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if create {
			if err := lockChannel(tx, input.GroupID); err != nil {
				return err
			}
		}
		if err := manager(tx, input.GroupID, actor); err != nil {
			return err
		}
		var user models.User
		if err := tx.Select("id", "name").First(&user, "id = ?", input.PublicID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return missing
			}
			return err
		}
		if user.ID == actor {
			return missing
		}
		relation, err := relationship(tx, actor, user.ID)
		if err != nil {
			return err
		}
		if relation == "blocked" {
			return missing
		}
		var count int64
		if err = tx.Model(&models.ChannelMember{}).Where("channel_id = ? AND user_id = ?", input.GroupID, user.ID).Count(&count).Error; err != nil {
			return err
		}
		state := "none"
		if count > 0 {
			state = "member"
		} else {
			if err = tx.Model(&models.ChannelInvite{}).Where("channel_id = ? AND invitee_id = ? AND status = 'pending' AND expires_at > ?", input.GroupID, user.ID, now).Count(&count).Error; err != nil {
				return err
			}
			if count > 0 {
				state = "invited"
			} else if relation == "accepted" {
				state = "connection"
			}
		}
		view = targetView{user.ID, user.ID.String(), user.Name, state}
		if !create {
			return nil
		}
		if state != "none" {
			return problem{409, "that user is already a member, invited, or can be added directly"}
		}
		if err = tx.Model(&models.ChannelInvite{}).Where("channel_id = ? AND invitee_id = ? AND status = 'declined' AND responded_at > ?", input.GroupID, user.ID, now.Add(-24*time.Hour)).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return problem{409, "please wait 24 hours after a declined invite"}
		}
		if err = tx.Model(&models.ChannelInvite{}).Where("channel_id = ? AND invitee_id = ? AND status = 'pending' AND expires_at <= ?", input.GroupID, user.ID, now).
			Updates(map[string]any{"status": "expired", "responded_at": now}).Error; err != nil {
			return err
		}
		return tx.Create(&models.ChannelInvite{ID: uuid.New(), ChannelID: input.GroupID, InviterID: actor, InviteeID: user.ID, Status: "pending", CreatedAt: now, ExpiresAt: now.Add(7 * 24 * time.Hour)}).Error
	})
	if err != nil {
		fail(w, err)
		return
	}
	if create {
		h.changed(r.Context(), input.GroupID, actor, view.ID)
		jsonReply(w, 201, view)
	} else {
		jsonReply(w, 200, view)
	}
}

func (h Handler) Inbox(w http.ResponseWriter, r *http.Request)   { h.list(w, r, false) }
func (h Handler) Pending(w http.ResponseWriter, r *http.Request) { h.list(w, r, true) }

func (h Handler) list(w http.ResponseWriter, r *http.Request, byChannel bool) {
	actor, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		fail(w, problem{401, "unauthorized"})
		return
	}
	tx := h.DB.WithContext(r.Context())
	query := tx.Table("channel_invites AS i").
		Select("i.id, i.channel_id, c.name AS channel_name, sender.name AS inviter_name, recipient.name AS invitee_name, i.invitee_id, i.status, i.expires_at").
		Joins("JOIN channels c ON c.id = i.channel_id").
		Joins("JOIN users sender ON sender.id = i.inviter_id").
		Joins("JOIN users recipient ON recipient.id = i.invitee_id").
		Joins("JOIN channel_members roles ON roles.channel_id = i.channel_id AND roles.user_id = i.inviter_id AND roles.role IN ('owner', 'admin')").
		Where("i.status = 'pending' AND i.expires_at > ?", time.Now().UTC())
	if byChannel {
		channel, parseErr := uuid.Parse(chi.URLParam(r, "channelID"))
		if parseErr != nil {
			fail(w, problem{400, "invalid channel"})
			return
		}
		if err = manager(tx, channel, actor); err != nil {
			fail(w, err)
			return
		}
		query = query.Where("i.channel_id = ?", channel)
	} else {
		query = query.Where("i.invitee_id = ?", actor)
	}
	rows := make([]inviteView, 0)
	if err = query.Order("i.created_at DESC").Scan(&rows).Error; err != nil {
		fail(w, err)
		return
	}
	jsonReply(w, 200, rows)
}

func (h Handler) Accept(w http.ResponseWriter, r *http.Request)  { h.transition(w, r, "accepted") }
func (h Handler) Decline(w http.ResponseWriter, r *http.Request) { h.transition(w, r, "declined") }
func (h Handler) Cancel(w http.ResponseWriter, r *http.Request)  { h.transition(w, r, "cancelled") }

func (h Handler) transition(w http.ResponseWriter, r *http.Request, status string) {
	actor, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		fail(w, problem{401, "unauthorized"})
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "inviteID"))
	if err != nil {
		fail(w, problem{400, "invalid invite"})
		return
	}
	var channel models.Channel
	var invite models.ChannelInvite
	err = h.DB.WithContext(r.Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&invite, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return missing
			}
			return err
		}
		if err := inviteResponsePermission(invite, actor, status); err != nil {
			return err
		}
		if err := lockChannel(tx, invite.ChannelID); err != nil {
			return err
		}
		if status == "cancelled" {
			if err := manager(tx, invite.ChannelID, actor); err != nil {
				return err
			}
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&invite, "id = ?", id).Error; err != nil {
			return err
		}
		now := time.Now().UTC()
		if err := pendingAt(invite, now); err != nil {
			return err
		}
		if status == "accepted" {
			if err := manager(tx, invite.ChannelID, invite.InviterID); err != nil {
				return closed
			}
			relation, err := relationship(tx, invite.InviterID, invite.InviteeID)
			if err != nil {
				return err
			}
			if relation == "blocked" {
				return closed
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models.ChannelMember{
				ChannelID: invite.ChannelID, UserID: actor, Role: models.RoleMember,
			}).Error; err != nil {
				return err
			}
			if err := tx.Select("id", "name").First(&channel, "id = ?", invite.ChannelID).Error; err != nil {
				return err
			}
		}
		return tx.Model(&invite).Updates(map[string]any{"status": status, "responded_at": now}).Error
	})
	if err != nil {
		fail(w, err)
		return
	}
	h.changed(r.Context(), invite.ChannelID, invite.InviterID, invite.InviteeID)
	if status == "accepted" {
		jsonReply(w, 200, map[string]any{"channel": map[string]any{"id": channel.ID, "name": channel.Name, "role": "member"}})
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func inviteResponsePermission(invite models.ChannelInvite, actor uuid.UUID, status string) error {
	if status != "cancelled" && actor != invite.InviteeID {
		return missing
	}
	return nil
}

func pendingAt(invite models.ChannelInvite, now time.Time) error {
	if invite.Status != "pending" {
		return closed
	}
	if !invite.ExpiresAt.After(now) {
		return expired
	}
	return nil
}

func lockChannel(tx *gorm.DB, id uuid.UUID) error {
	var channel models.Channel
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&channel, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return missing
		}
		return err
	}
	return nil
}
func manager(tx *gorm.DB, channel, actor uuid.UUID) error {
	var member models.ChannelMember
	err := tx.Clauses(clause.Locking{Strength: "SHARE"}).First(&member, "channel_id = ? AND user_id = ?", channel, actor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return denied
	}
	if err != nil {
		return err
	}
	if member.Role != models.RoleOwner && member.Role != models.RoleAdmin {
		return denied
	}
	return nil
}
func relationship(tx *gorm.DB, a, b uuid.UUID) (string, error) {
	var row models.Friend
	err := tx.Where("(user_low_id = ? AND user_high_id = ?) OR (user_low_id = ? AND user_high_id = ?)", a, b, b, a).Take(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	return row.Status, err
}
func readJSON(w http.ResponseWriter, r *http.Request, value any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2048))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return errors.New("unexpected JSON")
	}
	return nil
}
func fail(w http.ResponseWriter, err error) {
	var known problem
	if errors.As(err, &known) {
		http.Error(w, known.text, known.code)
		return
	}
	http.Error(w, "group invites are temporarily unavailable", 503)
}
func jsonReply(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func (h Handler) changed(ctx context.Context, channel uuid.UUID, parties ...uuid.UUID) {
	if h.Publish == nil {
		return
	}
	var managers []uuid.UUID
	_ = h.DB.WithContext(ctx).Model(&models.ChannelMember{}).
		Where("channel_id = ? AND role IN ?", channel, []string{"owner", "admin"}).
		Pluck("user_id", &managers).Error
	h.Publish(channel, append(parties, managers...)...)
}
