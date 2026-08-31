package handler

import (
	"time"

	"github.com/meteoradev/BlogApi/internal/infra/jwt"
	"github.com/meteoradev/BlogApi/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func NewRouter(
	rdb *redis.Client,
	userCtrl *UserController,
	postCtrl *PostController,
	authCtrl *AuthController,
	secret string,
	expiry time.Duration,
	PublicRPM int,
	ProtectedPRM int,
	logger zerolog.Logger,
) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.LoggingMiddleware(logger))

	// Public routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.NewRateLimitMiddleware(rdb, PublicRPM).Limit)
		r.Post("/register", userCtrl.Create)
		r.Post("/login", authCtrl.Login)
		r.Get("/posts/id/{postID}", postCtrl.GetByID)
		r.Get("/posts/title/{title}", postCtrl.GetByTitle)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.NewRateLimitMiddleware(rdb, ProtectedPRM).Limit)
		r.Use(middleware.NewAuthMiddleware(jwt.NewProvider(secret, expiry)).Authenticate)

		r.Route("/users", func(r chi.Router) {
			r.Get("/id/{userID}", userCtrl.GetByID)
			r.Get("/email/{email}", userCtrl.GetByEmail)
			r.Put("/{userID}", userCtrl.Update)
			r.Delete("/{userID}", userCtrl.Delete)
		})
		r.Route("/posts", func(r chi.Router) {
			r.Post("/", postCtrl.Create)
			r.Put("/{postID}", postCtrl.Update)
			r.Delete("/{postID}", postCtrl.Delete)
		})
	})
	return r
}
