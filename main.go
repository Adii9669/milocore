package main

import (
	"log"
	"net/http"

	//my packages
	"chat-server/internals/config"
	"chat-server/internals/db"
	"chat-server/internals/handlers/auth"
	"chat-server/internals/handlers/emails"
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

	//2. Create a websockets instance and run it
	hub := websockets.NewHub()
	go hub.Run()

	//3. Set Up the Router
	r := mux.NewRouter()

	api := r.PathPrefix("/api").Subrouter()

	apiRouter := api.PathPrefix("/auth").Subrouter()
	//Public routes
	apiRouter.HandleFunc("/register", auth.RegisterHandler).Methods("POST")
	apiRouter.HandleFunc("/login", auth.LoginHandler).Methods("POST")
	apiRouter.HandleFunc("/verify", emails.VerifyEmailHandler).Methods("GET")
	apiRouter.HandleFunc("/resend-verification", emails.ResendVerificationHandler).Methods("POST")

	//Protected Routes
	secureRouter := api.PathPrefix("/").Subrouter()
	secureRouter.Use(middleware.MuxAdapter(middleware.AuthMiddleware()))

	//protecting the websockets routes
	//now the api will check it throught AuthMiddleware
	secureRouter.HandleFunc("/me", auth.MeHandler).Methods("GET")
	secureRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
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
