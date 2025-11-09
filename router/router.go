package router

import (
	"log"
	"net/http"

	//my packages
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

func SetUpRouter(userRepo repository.UserRepository, crewRepo repository.CrewRepository, hub *websockets.Hub) http.Handler {

	//1. Set Up the Router
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
	protectedRouter.HandleFunc("/crews/{id}", crews.DeleteCrewHandler(crewRepo)).Methods("DELETE")
	protectedRouter.HandleFunc("/me", auth.MeHandler).Methods("GET")
	protectedRouter.HandleFunc("/logout", auth.LogoutHandler).Methods("POST")
	protectedRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websockets.ServeWs(hub, w, r)
	})

	// DEBUG: show all registered routes
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		log.Printf("ROUTE: %v %v", methods, path)
		return nil
	})

	// CORS Configuration
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	allowedCredentials := handlers.AllowCredentials()

	return handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders, allowedCredentials)(r)

}
