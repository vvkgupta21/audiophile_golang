package server

import (
	"audio_phile/database/handler"
	"github.com/go-chi/chi/v5"
)

func UserRoute(r chi.Router) {
	r.Group(func(user chi.Router) {
		user.Route("/product", func(product chi.Router) {
			product.Get("/", handler.GetAllProduct)
			product.Get("/{id}", handler.GetProductById)
		})
		user.Route("/address", func(address chi.Router) {
			address.Post("/", handler.CreatedAddress)
			address.Get("/", handler.GetUserAddress)
		})
		user.Route("/cart", func(cartProduct chi.Router) {
			cartProduct.Post("/{id}/{quantity}", handler.CreateProductToCart)
			cartProduct.Get("/", handler.GetCartWithProductById)
			cartProduct.Route("/add", func(addQuantity chi.Router) {
				addQuantity.Post("/{cartId}/{productId}", handler.AddProductQuantityInCart)
			})
			cartProduct.Route("/remove", func(removeQuantity chi.Router) {
				removeQuantity.Post("/{cartId}/{productId}", handler.RemoveProductQuantityInCart)
			})
			cartProduct.Route("/delete", func(deleteProduct chi.Router) {
				deleteProduct.Delete("/{cartId}/{productId}", handler.DeleteProductFromCart)
			})
		})
		user.Route("/order", func(order chi.Router) {
			order.Post("/{cartId}/{addressId}", handler.CreateOrder)
			//order.Post("/{productId}/{quantity}", handler.UpdateProductQuantity)
			//order.Get("/{cartId}", handler.GetCartProductIds)
		})
	})
}
