package main

import (
	"log"
	"net/http"
	"os"

	//my packages
	"chat-server/internals/config"
	"chat-server/internals/db"
	"chat-server/internals/handlers/auth"
	"chat-server/internals/handlers/emails"
	"chat-server/internals/websockets"
	"chat-server/middleware"

	// go libraries
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {

	//Load Env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error Loading Env file.")
	}

	//Load the config
	if err := config.LoadConfig(); err != nil {
		log.Printf("Error loading Config file %v", err)
	}

	log.Println("--- CONFIGURATION DIAGNOSTIC (from main.go) ---")
	log.Printf("SMTP_HOST at startup: '%s'", os.Getenv("SMTP_HOST"))
	log.Printf("SMTP_PORT at startup: '%s'", os.Getenv("SMTP_PORT"))
	log.Printf("SMTP_USER at startup: '%s'", os.Getenv("SMTP_USER"))
	log.Printf("SMTP_USER at startup: '%s'", os.Getenv("EMAIL_FROM"))
	log.Println("---------------------------------------------")

	log.Printf("SMTP_HOST at startup: '%s'", config.Cfg.Email.SMTPHost)
	log.Printf("SMTP_PORT at startup: '%s'", config.Cfg.Email.SMTPPort)
	log.Printf("SMTP_USER at startup: '%s'", config.Cfg.Email.SMTPUser)
	log.Printf("SMTP_USER at startup: '%s'", config.Cfg.Email.EmailFrom)
	log.Println("---------------------------------------------")

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

	port := ":8000"
	log.Printf("Server is running on http://localhost%s\n", port)
	if err := http.ListenAndServe(port, corsHandler(r)); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
