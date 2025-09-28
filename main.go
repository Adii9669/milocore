package main

import (
	"log"
	"net/http"

	//my packages
	"chat-server/internals/config"
	"chat-server/internals/db"
	"chat-server/internals/handlers/auth"
	"chat-server/internals/handlers/crews"
	"chat-server/internals/handlers/emails"
	"chat-server/internals/repository"
	"chat-server/internals/websockets"
	"chat-server/middleware"

	// go libraries
	"github.com/gorilla/handlers" //for the websockets connection
	"github.com/gorilla/mux"      //Mux for the routing of private and public pages (user)
)

func main() {

	//Load the config
	if err := config.LoadConfig(); err != nil {
		log.Printf("Error loading Config file %v", err)
	}

	//connecting to the database
	db.ConnectToDB()

	//intializing the repo
	userRepo := &repository.GormUserRepository{}
	crewRepo := &repository.GormCrewRepository{}

	//2. Create a websockets instance and run it
	hub := websockets.NewHub()
	go hub.Run()

	//3. Set Up the Router
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	apiRouter := api.PathPrefix("/auth").Subrouter()
	//Public routes
	apiRouter.HandleFunc("/register", auth.RegisterHandler(userRepo)).Methods("POST")
	apiRouter.HandleFunc("/login", auth.LoginHandler(userRepo)).Methods("POST")
	apiRouter.HandleFunc("/verify-otp", auth.VerifyOtpHandler(userRepo)).Methods("POST")
	apiRouter.HandleFunc("/resend-verification", emails.ResendVerificationHandler).Methods("POST")
	apiRouter.HandleFunc("/check-availability", auth.CheckAvailablityHandler).Methods("POST")

	//Protected Routes
	protectedRouter := r.PathPrefix("/").Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware)

	//use
	protectedRouter.HandleFunc("/crews", crews.CreateCrewHandler(crewRepo)).Methods("POST")
	protectedRouter.HandleFunc("/crews", crews.Getcrew(crewRepo)).Methods("GET")
	protectedRouter.HandleFunc("/me", auth.MeHandler).Methods("GET")
	protectedRouter.HandleFunc("/logout", auth.LogoutHandler).Methods("POST")
	protectedRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websockets.ServeWs(hub, w, r)
	})

	// CORS Configuration
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	allowedCredentials := handlers.AllowCredentials()

	corsHandler := handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders, allowedCredentials)

	// 6. Start the server USING the loaded configuration
	port := ":" + config.Cfg.Server.PORT
	log.Printf("Server is running on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, corsHandler(r)); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
