package router

import (
	"getitdone/internal/handler"
	"getitdone/internal/middleware"
	"getitdone/internal/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func New(
	authHandler *handler.AuthHandler,
	jwtService *service.JWTService,
	tokenBlacklist *service.TokenBlacklist,
	familyHandler *handler.FamilyHandler,
) *chi.Mux {
	r := chi.NewRouter()
	authMiddleware := middleware.AuthMiddleware(
		jwtService,
		tokenBlacklist,
	)

	r.Use(chimiddleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"OK"}`))
	})

	r.Post("/api/v1/auth/register", authHandler.Register)
	r.Post("/api/v1/auth/login", authHandler.Login)
	r.Post("/api/v1/auth/refresh", authHandler.Refresh)
	r.With(authMiddleware).Post("/api/v1/auth/logout", authHandler.Logout)
	r.With(authMiddleware).Get("/api/v1/profile", profileHandler)
	r.With(authMiddleware).Post(
		"/api/v1/families",
		familyHandler.Create,
	)
	r.With(authMiddleware).Get(
		"/api/v1/families/{id}",
		familyHandler.GetByID)
	r.With(authMiddleware).Get(
		"/api/v1/families",
		familyHandler.GetFamilies,
	)
	r.With(authMiddleware).Patch(
		"/api/v1/families/{id}",
		familyHandler.Update,
	)
	r.With(authMiddleware).Delete(
		"/api/v1/families/{id}",
		familyHandler.Delete,
	)

	return r
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserID(r)

	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(
		`{"user_id":"` + userID.String() + `"}`,
	))
}
