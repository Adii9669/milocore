package router

import (
	"chat-server/internals/app"
	"chat-server/internals/handlers/auth"
	"chat-server/internals/handlers/chathistory"
	"chat-server/internals/handlers/crews"
	"chat-server/internals/handlers/friends"
	"chat-server/internals/websockets"
	"chat-server/middleware"
	"log"

	// go libraries
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/handlers" //for the websockets connection
	// "github.com/gorilla/mux"      //Mux for the routing of private and public pages (user)
)

func SetUpRouter(app *app.App) http.Handler {

	//1. Set Up the Router
	// r := mux.NewRouter()
	r := chi.NewRouter()

	/* ---------------------------------------------------------
	   GLOBAL MIDDLEWARE
	--------------------------------------------------------- */
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	/* ---------------------------------------------------------
	   PUBLIC ROUTES (NO AUTH)
	--------------------------------------------------------- */

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth.RegisterHandler(app.UserRepo))
		r.Post("/login", auth.LoginHandler(app.UserRepo))
		r.Post("/verify-otp", auth.VerifyOtpHandler(app.UserRepo))
		r.Post("/check-availability", auth.CheckAvailablityHandler)

	})

	/* ---------------------------------------------------------
	   PROTECTED ROUTES (AUTH REQUIRED)
	--------------------------------------------------------- */

	// handlers for creating the instace for the repository for using
	crewHandler := crews.NewCrewHandler(app.CrewRepo)
	authHandler := auth.NewAuthHandler(app.UserRepo)
	frndHandler := friends.NewFriendHandler(app.FriendRepo, authHandler.UserRepo)
	chathistoryHandler := chathistory.NewHandler(app.ChatHistoryService)

	//Protected Routes
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		// Crew
		r.Post("/crews", crewHandler.CreateCrew)
		r.Get("/crews", crewHandler.Getcrew)
		r.Delete("/crew/{id}", crewHandler.DeleteCrew)

		// User
		r.Get("/me", authHandler.Me)

		// Friend
		r.Post("/friend/request", frndHandler.SendFrndRequest)
		r.Post("/friend/accept", frndHandler.AcceptRequest)
		r.Get("/friends", frndHandler.GetFriends)

		// Chat
		r.Get("/chats/crew/{crewId}", chathistoryHandler.GetCrewHistory)
		r.Get("/chats/dm/{userId}", chathistoryHandler.GetDmHistory)

		// r.Poat
		// WebSocket
		r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
			websockets.ServeWS(app.Hub, app.MessageService, w, r)
		})

		// Logout
		r.Post("/logout", auth.LogoutHandler)
	})

	// DEBUG: show all registered routes
	// r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
	// 	path, _ := route.GetPathTemplate()
	// 	methods, _ := route.GetMethods()
	// 	log.Printf("ROUTE: %v %v", methods, path)
	// 	return nil
	// })
	chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		log.Printf("%s %s\n", method, route)
		return nil
	})

	// CORS Configuration
	/* ---------------------------------------------------------
	   CORS
	--------------------------------------------------------- */
	return handlers.CORS(
		handlers.AllowedOrigins([]string{
			"http://localhost:3000",
			"https://milonext.onrender.com",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "content-type", "Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)(r)

}
