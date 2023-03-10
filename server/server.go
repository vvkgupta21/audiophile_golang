package server

import (
	"audio_phile/database/handler"
	"audio_phile/middleware"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

const (
	readTimeout       = 5 * time.Minute
	readHeaderTimeout = 30 * time.Second
	writeTimeout      = 5 * time.Minute
)

type Server struct {
	chi.Router
	server *http.Server
}

func SetupRoutes() *Server {
	routes := chi.NewRouter()
	routes.Route("/api", func(api chi.Router) {
		api.Post("/register", handler.CreateUser)
		api.Post("/login", handler.Login)
		api.Route("/admin", func(admin chi.Router) {
			admin.Use(middleware.AuthMiddleware)
			admin.Use(middleware.AdminMiddleware)
			admin.Group(AdminRoute)

		})
		api.Route("/user", func(user chi.Router) {
			user.Use(middleware.AuthMiddleware)
			user.Use(middleware.UserMiddleware)
			user.Group(UserRoute)
		})
	})
	return &Server{
		Router: routes,
	}
}

func (srv *Server) Run(port string) error {
	srv.server = &http.Server{
		Addr:              port,
		Handler:           srv.Router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}
	return srv.server.ListenAndServe()
}
