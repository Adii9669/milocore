package router

import (
	"log"
	"net/http"

	//my packages
	"chat-server/internals/app"
	"chat-server/internals/handlers/auth"
	"chat-server/internals/handlers/crews"
	"chat-server/internals/handlers/friends"
	"chat-server/internals/handlers/messages"
	"chat-server/internals/websockets"
	"chat-server/middleware"

	// go libraries
	"github.com/gorilla/handlers" //for the websockets connection
	"github.com/gorilla/mux"      //Mux for the routing of private and public pages (user)
)

func SetUpRouter(app *app.App) http.Handler {

	//1. Set Up the Router
	r := mux.NewRouter()

	/* ---------------------------------------------------------
	   PUBLIC ROUTES (NO AUTH)
	--------------------------------------------------------- */

	api := r.PathPrefix("/api").Subrouter()
	apiRouter := api.PathPrefix("/auth").Subrouter()

	apiRouter.HandleFunc("/register", auth.RegisterHandler(app.UserRepo)).Methods("POST")
	apiRouter.HandleFunc("/login", auth.LoginHandler(app.UserRepo)).Methods("POST")
	apiRouter.HandleFunc("/verify-otp", auth.VerifyOtpHandler(app.UserRepo)).Methods("POST")
	apiRouter.HandleFunc("/check-availability", auth.CheckAvailablityHandler).Methods("POST")

	/* ---------------------------------------------------------
	   PROTECTED ROUTES (AUTH REQUIRED)
	--------------------------------------------------------- */

	//Protected Routes
	protectedRouter := r.PathPrefix("/").Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware)

	// Crew handlers
	crewHandler := crews.NewCrewHandler(app.CrewRepo)
	authHandler := auth.NewAuthHandler(app.UserRepo)

	protectedRouter.HandleFunc("/crews", crewHandler.CreateCrew).Methods("POST")
	protectedRouter.HandleFunc("/crews", crewHandler.Getcrew).Methods("GET")
	protectedRouter.HandleFunc("/crews/{id}", crewHandler.DeleteCrew).Methods("DELETE")
	protectedRouter.HandleFunc("/me", authHandler.Me).Methods("GET")

	//repo
	// userRepo :
	frndHandler := friends.NewFriendHandler(app.FriendRepo, authHandler.UserRepo)

	//Friend
	protectedRouter.HandleFunc("/friend/request", frndHandler.SendFrndRequest).Methods("POST")
	protectedRouter.HandleFunc("/friend/accept", frndHandler.AcceptRequest).Methods("POST")

	// Message handlers
	messageHandler := messages.NewMessageHandler(app.MessageRepo)
	protectedRouter.HandleFunc("/get-messages", messageHandler.GetMessage).Methods("GET")

	//logout
	protectedRouter.HandleFunc("/logout", auth.LogoutHandler).Methods("POST")

	//websockets route
	protectedRouter.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websockets.ServeWs(app.Hub, w, r)
	})

	// DEBUG: show all registered rout
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		path, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()
		log.Printf("ROUTE: %v %v", methods, path)
		return nil
	})

	// CORS Configuration
	allowedOrigins := handlers.AllowedOrigins([]string{
		"http://localhost:3000",
		"https://milonext.onrender.com",
	})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	allowedCredentials := handlers.AllowCredentials()

	return handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders, allowedCredentials)(r)

}
