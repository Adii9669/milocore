package main

import (
	"log"
	"net/http"
	"time"

	//my packages
	"chat-server/internals/app"
	"chat-server/internals/config"
	"chat-server/internals/db"
	"chat-server/internals/utils"
	"chat-server/internals/websockets"
	"chat-server/router"
)

func StartCleanupScheduler() {
	go func() {
		for {
			utils.CleanupUnverifiedUsers()
			time.Sleep(24 * time.Hour) // run once a day
		}
	}()
}

func main() {

	//Load the config
	if err := config.LoadConfig(); err != nil {
		log.Printf("Error loading Config file %v", err)
	}

	// log.Println("API", config.Cfg.Email.ResendApi)
	// log.Println("EMAIL_FROM:", config.Cfg.Email.EmailFrom)
	//connecting to the database
	database := db.ConnectToDB()

	//3.clean UP
	StartCleanupScheduler()

	//4. Create a websockets instance and run it
	hub := websockets.NewHub()
	go hub.Run()

	cfg := &config.Cfg

	//5.create app container (DI)
	application := app.NewApp(database, hub, cfg)
	r := router.SetUpRouter(application)

	// 6. Start the server USING the loaded configuration
	port := ":" + config.Cfg.Server.PORT
	log.Printf("Server is running on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {

		log.Fatal("ListenAndServe:", err)
	}
}
