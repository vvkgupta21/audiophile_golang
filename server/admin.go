package server

import (
	"audio_phile/database/handler"
	"github.com/go-chi/chi/v5"
)

func AdminRoute(r chi.Router) {
	r.Group(func(admin chi.Router) {
		admin.Route("/product", func(product chi.Router) {
			product.Post("/", handler.CreateProduct)
			product.Get("/", handler.GetAllProduct)
			product.Get("/{id}", handler.GetProductById)
		})
		admin.Route("/user", func(user chi.Router) {
			user.Get("/", handler.GetAllUser)
			user.Get("/{id}", handler.GetUserByUserId)
			user.Delete("/{id}", handler.DeleteUserByUserId)
		})
		admin.Route("/status", func(orderStatus chi.Router) {
			orderStatus.Post("/{orderId}/{orderStatus}", handler.CreateOrderStatus)
		})
	})
}
