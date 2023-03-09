package server

import (
	"audio_phile/database/handler"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	readTimeout       = 5 * time.Minute
	readHeaderTimeout = 30 * time.Second
	writeTimeout      = 5 * time.Minute
)

type Server struct {
	*gin.Engine
	server *http.Server
}

func SetupRoutes() *Server {
	//routes := chi.NewRouter()

	routes := gin.Default()

	api := routes.Group("/api")
	{
		api.POST("/register", handler.CreateUser)
		api.POST("/login", handler.Login)

	}

	//routes.Group("/api")
	//
	//api := routes.POST("/login")
	//{
	//
	//}
	//	routes.Route("/api", func(api chi.Router) {
	//	api.Post("/register", handler.CreateUser)
	//	api.Post("/login", handler.Login)
	//	api.Route("/admin", func(admin chi.Router) {
	//		admin.Use(middleware.AuthMiddleware)
	//		admin.Use(middleware.AdminMiddleware)
	//		admin.Group(AdminRoute)
	//
	//	})
	//	api.Route("/user", func(user chi.Router) {
	//		user.Use(middleware.AuthMiddleware)
	//		user.Use(middleware.UserMiddleware)
	//		user.Group(UserRoute)
	//	})
	//})
	return &Server{
		Engine: routes,
	}
}

func (srv *Server) Run(port string) error {
	srv.server = &http.Server{
		Addr:              port,
		Handler:           srv.Engine,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
	}
	return srv.server.ListenAndServe()
}

func (srv *Server) Stop(time time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time)
	defer cancel()
	return srv.server.Shutdown(ctx)
}
