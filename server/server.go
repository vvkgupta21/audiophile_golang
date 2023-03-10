package server

import (
	"audio_phile/database/handler"
	"audio_phile/middleware"
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
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminMiddleware())
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

		user := api.Group("/user")
		user.Use(middleware.AuthMiddleware())
		user.Use(middleware.UserMiddleware())
		{
			product := user.Group("/product")
			{
				product.GET("/", handler.GetAllProduct)
				product.GET("/:id", handler.GetProductById)
			}
			address := user.Group("/address")
			{
				address.POST("/", handler.CreatedAddress)
				address.GET("/", handler.GetUserAddress)
			}
			cart := user.Group("/cart")
			{
				cart.POST("/:id/:quantity", handler.CreateProductToCart)
				cart.GET("/", handler.GetCartWithProductById)

				addQuantity := cart.Group("/add")
				{
					addQuantity.POST("/:cartId/:productId", handler.AddProductQuantityInCart)
				}
				removeQuantity := cart.Group("/remove")
				{
					removeQuantity.POST("/:cartId/:productId", handler.RemoveProductQuantityInCart)
				}
				deleteProduct := cart.Group("/delete")
				{
					deleteProduct.DELETE("/:cartId/:productId", handler.DeleteProductFromCart)
				}
				order := cart.Group("/order")
				{
					order.POST("/:cartId/:addressId", handler.CreateOrder)
				}
			}
		}
	}
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
