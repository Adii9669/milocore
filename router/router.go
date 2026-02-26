package router

import (
	"chat-server/internals/app"
	"chat-server/internals/handlers/auth"
	"chat-server/internals/handlers/chathistory"
	"chat-server/internals/handlers/crews"
	"chat-server/internals/handlers/friends"
	"chat-server/internals/middleware"
	"chat-server/internals/websockets"
	"log"
	"os"

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
	crewHandler := crews.NewCrewHandler(app.CrewService)
	authHandler := auth.NewAuthHandler(app.UserRepo)
	frndHandler := friends.NewFriendHandler(app.FriendService)

	chathistoryHandler := chathistory.NewHandler(app.ChatHistoryService)

	//Protected Routes
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		// Crew
		r.Post("/crew/create", crewHandler.CreateCrew)
		r.Get("/crews", crewHandler.Getcrew)
		r.Delete("/crew/{crewID}", crewHandler.DeleteCrew)
		r.Put("/crew/{crewID}", crewHandler.RenameCrew)

		//member
		r.Get("/{crewID}/members", crewHandler.GetMembers)
		r.Post("/crew/{crewID}/member/{memberID}", crewHandler.AddMember)
		r.Delete("/crew/{crewID}/member/{memberID}", crewHandler.RemoveMember)
		r.Put("/crew/{crewID}/member/{memberID}/role", crewHandler.UpdateRole)

		// User
		r.Get("/me", authHandler.Me)

		// Friend
		r.Post("/friend/request", frndHandler.SendFrndRequest)
		r.Post("/friend/accept", frndHandler.AcceptRequest)
		r.Get("/friends", frndHandler.GetFriends)
		r.Delete("/friend/remove/{friendID}", frndHandler.RemoveFriend)

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

	if os.Getenv("APP_ENV") == "local" {
		chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
			log.Printf("%s %s\n", method, route)
			return nil
		})
	}

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
