package api

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/netmap/netmap/internal/api/handlers"
	"github.com/netmap/netmap/internal/api/middleware"
	"github.com/netmap/netmap/internal/api/ws"
	"github.com/netmap/netmap/internal/store"
)

func NewRouter(
	s *store.Store,
	hub *ws.Hub,
	scanHandler *handlers.ScanHandler,
	configHandler *handlers.ConfigHandler,
	authHandler *handlers.AuthHandler,
	alertHandler *handlers.AlertHandler,
	version string,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(middleware.CORS()))

	devices := handlers.NewDeviceHandler(s.Devices)
	networks := handlers.NewNetworkHandler(s.Networks)
	scans := scanHandler
	system := handlers.NewSystemHandler(s.Devices, version)

	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes — always public
		r.Route("/auth", func(r chi.Router) {
			r.Get("/status", authHandler.Status)
			r.Post("/setup", authHandler.Setup)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
		})

		// All other API routes require authentication (when password is set up)
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(s.Sessions, authHandler))

			r.Route("/devices", func(r chi.Router) {
				r.Get("/", devices.List)
				r.Post("/", devices.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", devices.Get)
					r.Put("/", devices.Update)
					r.Delete("/", devices.Delete)
				})
			})

			r.Route("/networks", func(r chi.Router) {
				r.Get("/", networks.List)
				r.Post("/", networks.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", networks.Get)
					r.Put("/", networks.Update)
					r.Delete("/", networks.Delete)
				})
			})

			r.Route("/scans", func(r chi.Router) {
				r.Get("/", scans.List)
				r.Post("/", scans.Trigger)
				r.Get("/{id}", scans.Get)
				r.Delete("/{id}", scans.Cancel)
			})

			r.Route("/alerts", func(r chi.Router) {
				r.Get("/", alertHandler.List)
				r.Post("/", alertHandler.Create)
				r.Post("/read", alertHandler.MarkRead)
				r.Delete("/", alertHandler.DeleteAll)
			})

			r.Get("/system/status", system.Status)
			r.Get("/system/config", configHandler.Get)
			r.Put("/system/config", configHandler.Put)

			r.Get("/ws", hub.HandleWS)
		})
	})

	return r
}
