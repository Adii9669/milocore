package router

import (
	"chat-server/internals/app"
	"chat-server/internals/handlers/auth"
	"chat-server/internals/handlers/channels"
	"chat-server/internals/handlers/chathistory"
	"chat-server/internals/handlers/files"
	"chat-server/internals/handlers/friends"
	"chat-server/internals/handlers/messages"
	"chat-server/internals/handlers/plans"
	"chat-server/internals/middleware"
	"chat-server/internals/websockets"
	"encoding/json"
	"log"

	// go libraries
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/handlers" //for the websockets connection

	// "github.com/gorilla/mux"      //Mux for the routing of private and public pages (user)
	httpSwagger "github.com/swaggo/http-swagger"
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

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := app.DB.WithContext(r.Context()).Exec("SELECT 1").Error; err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			return
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", auth.RegisterHandler(app.AuthService))
		r.Post("/login", auth.LoginHandler(app.AuthService))
		r.Post("/verify-otp", auth.VerifyOtpHandler(app.AuthService))
		r.Post("/resend-otp", auth.ResendOTPHandler(app.AuthService))
		r.Post("/check-availability", auth.CheckAvailablityHandler)
		r.Post("/refresh", auth.RefreshHandler(app.AuthService))

	})

	/* ---------------------------------------------------------
	   PROTECTED ROUTES (AUTH REQUIRED)
	--------------------------------------------------------- */

	// handlers for creating the instace for the repository for using
	channelHandler := channels.NewChannelHandler(app.ChannelService)
	authHandler := auth.NewAuthHandler(app.UserRepo)
	frndHandler := friends.NewFriendHandler(app.FriendService)
	chathistoryHandler := chathistory.NewHandler(app.ChatHistoryService)
	fileHandler := files.NewFileHandler(app.FileStorage, app.MessageService, app.Hub)
	messageHandler := messages.NewHandler(app.MessageService, app.FileStorage, app.Hub)
	planHandler := plans.NewHandler(app.PlanService)

	//Protected Routes
	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		r.Route("/channels", func(r chi.Router) {
			r.Post("/", channelHandler.CreateChannel)
			r.Get("/", channelHandler.GetChannels)

			r.Route("/{channelID}", func(r chi.Router) {
				r.Delete("/", channelHandler.DeleteChannel)
				r.Put("/", channelHandler.RenameChannel)

				r.Route("/members", func(r chi.Router) {
					r.Get("/", channelHandler.GetMembers)
					r.Post("/{userID}", channelHandler.AddMember)

					r.Route("/{userID}", func(r chi.Router) {
						r.Put("/role", channelHandler.UpdateMember)
						r.Delete("/", channelHandler.RemoveMember)
					})

				})
			})
		})

		// User
		r.Get("/me", authHandler.Me)
		r.Get("/users", authHandler.GetUsers)

		// Friend
		r.Get("/friends", frndHandler.GetFriends)

		r.Route("/friend-requests", func(r chi.Router) {
			r.Post("/", frndHandler.SendFrndRequest)
			r.Get("/", frndHandler.GetRequests)

			r.Route("/{requestID}", func(r chi.Router) {
				r.Put("/accept", frndHandler.AcceptRequest)
				r.Delete("/", frndHandler.RejectRequest)
			})
		})

		r.Delete("/friends/{friendID}", frndHandler.RemoveFriend)

		// Chat
		r.Delete("/messages/{messageID}", messageHandler.Delete)
		r.Post("/chats/channel/{channelId}/plans", planHandler.CreateChannelPlan)
		r.Get("/chats/channel/{channelId}/plans", planHandler.ListChannelPlans)
		r.Post("/chats/dm/{userId}/plans", planHandler.CreateDMPlan)
		r.Get("/chats/dm/{userId}/plans", planHandler.ListDMPlans)
		r.Get("/plans/{planId}", planHandler.GetPlan)
		r.Get("/plans/{planId}/poll", planHandler.GetPoll)
		r.Post("/plans/{planId}/poll", planHandler.CreatePoll)
		r.Put("/plans/{planId}/poll/votes", planHandler.SetPollVotes)
		r.Put("/plans/{planId}", planHandler.UpdatePlan)
		r.Put("/plans/{planId}/status", planHandler.UpdatePlanStatus)
		r.Put("/plans/{planId}/response", planHandler.SetResponse)
		r.Get("/plans/{planId}/participants", planHandler.GetParticipants)
		r.Get("/chats/channel/{channelId}", chathistoryHandler.GetChannelHistory)
		r.Get("/chats/dm/{userId}", chathistoryHandler.GetDmHistory)

		// r.Poat
		// WebSocket
		r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
			websockets.ServeWS(app.Hub, app.MessageService, w, r)
		})

		// Logout
		r.Post("/logout", auth.LogoutHandler(app.AuthService))

		//files
		r.Post("/files/upload", fileHandler.Upload)
		r.Get("/files/{messageID}/download", fileHandler.Download)
		r.Get("/files/{messageID}/preview", fileHandler.Preview)
		// r.Delete("/files/{objectName}", fileHandler.Delete)
	})

	// DEBUG: show all registered routes
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
			"https://milonext.vercel.app",
			"https://milonext.onrender.com",
		}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"X-Requested-With", "content-type", "Content-Type", "Authorization"}),
		handlers.AllowCredentials(),
	)(r)

}
