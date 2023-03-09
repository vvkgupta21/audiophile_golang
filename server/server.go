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
	routes := gin.Default()
	api := routes.Group("/api")
	{
		api.POST("/register", handler.CreateUser)
		api.POST("/login", handler.Login)

		admin := api.Group("/admin")
		{
			product := admin.Group("/product")
			{
				product.POST("/", handler.CreateProduct)
				product.GET("/", handler.GetAllProduct)
				product.GET("/:id", handler.GetProductById)
			}

			users := admin.Group("/user")
			{
				users.GET("/", handler.GetAllUser)
				users.GET("/:id", handler.GetUserByUserId)
				users.DELETE("/:id", handler.DeleteUserByUserId)
			}

			orderStatus := admin.Group("/status")
			{
				orderStatus.POST("/:orderId/:orderStatus", handler.CreateOrderStatus)
			}

			image := admin.Group("/image")
			{
				image.POST("/:productID", handler.UploadImages)
				image.GET("/:productID", handler.GetAllImageByProductId)
			}

		}

	}

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
