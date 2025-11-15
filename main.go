package main

import (
	"log"
	"net/http"
	"time"

	//my packages
	"chat-server/internals/config"
	"chat-server/internals/db"
	"chat-server/internals/repository"
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

	//connecting to the database
	db.ConnectToDB()

	StartCleanupScheduler()
	//intializing the repo
	//this is for the direct use when don't want any dependnicy injection in your program
	// userRepo := &repository.GormUserRepository{}
	// crewRepo := &repository.GormCrewRepository{}
	var userRepo repository.UserRepository = &repository.GormUserRepository{}
	var crewRepo repository.CrewRepository = &repository.GormCrewRepository{}

	//2. Create a websockets instance and run it
	hub := websockets.NewHub()
	go hub.Run()

	r := router.SetUpRouter(userRepo, crewRepo, hub)

	// 3. Start the server USING the loaded configuration
	port := ":" + config.Cfg.Server.PORT
	log.Printf("Server is running on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, r); err != nil {

		log.Fatal("ListenAndServe:", err)
	}
}
