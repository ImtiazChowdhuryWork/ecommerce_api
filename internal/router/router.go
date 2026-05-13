package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"ecommerce-api/internal/config"
	"ecommerce-api/internal/handlers"
	"ecommerce-api/internal/middleware"
	"ecommerce-api/internal/repository"
)

func New(
	cfg *config.Config,
	userRepo repository.UserRepository,
	categoryRepo repository.CategoryRepository,
	productRepo repository.ProductRepository,
	cartRepo repository.CartRepository,
	orderRepo repository.OrderRepository,
	reviewRepo repository.ReviewRepository,
) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.CORS)

	// Init handlers
	authH := handlers.NewAuthHandler(userRepo, cfg)
	userH := handlers.NewUserHandler(userRepo)
	catH := handlers.NewCategoryHandler(categoryRepo)
	productH := handlers.NewProductHandler(productRepo)
	cartH := handlers.NewCartHandler(cartRepo, productRepo)
	orderH := handlers.NewOrderHandler(orderRepo, cartRepo)
	reviewH := handlers.NewReviewHandler(reviewRepo, productRepo)

	auth := middleware.NewAuth(cfg)

	r.Route("/api/v1", func(r chi.Router) {
		// Health check
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"ok"}`))
		})

		// Auth — public
		r.Post("/auth/register", authH.Register)
		r.Post("/auth/login", authH.Login)
		r.Post("/auth/refresh", authH.Refresh)

		// Categories — list & detail public, CRUD admin-only
		r.Get("/categories", catH.List)
		r.Get("/categories/{id}", catH.GetByID)

		// Products — list & detail public, CRUD admin-only
		r.Get("/products", productH.List)
		r.Get("/products/{id}", productH.GetByID)
		r.Get("/products/{id}/reviews", reviewH.ListByProduct)

		// Authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(auth.Authenticate)

			// User profile
			r.Get("/users/me", userH.GetMe)
			r.Put("/users/me", userH.UpdateMe)
			r.Put("/users/me/password", userH.ChangePassword)

			// Cart
			r.Get("/cart", cartH.GetCart)
			r.Post("/cart/items", cartH.AddItem)
			r.Put("/cart/items/{id}", cartH.UpdateItem)
			r.Delete("/cart/items/{id}", cartH.RemoveItem)
			r.Delete("/cart", cartH.ClearCart)

			// Orders
			r.Post("/orders", orderH.Create)
			r.Get("/orders", orderH.List)
			r.Get("/orders/{id}", orderH.GetByID)
			r.Post("/orders/{id}/cancel", orderH.Cancel)

			// Reviews
			r.Post("/products/{id}/reviews", reviewH.Create)
			r.Put("/reviews/{id}", reviewH.Update)
			r.Delete("/reviews/{id}", reviewH.Delete)

			// Admin-only
			r.Group(func(r chi.Router) {
				r.Use(auth.RequireAdmin)

				r.Post("/categories", catH.Create)
				r.Put("/categories/{id}", catH.Update)
				r.Delete("/categories/{id}", catH.Delete)

				r.Post("/products", productH.Create)
				r.Put("/products/{id}", productH.Update)
				r.Delete("/products/{id}", productH.Delete)

				r.Put("/orders/{id}/status", orderH.UpdateStatus)
			})
		})
	})

	return r
}
