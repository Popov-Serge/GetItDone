package middleware

import (
	"context"
	"getitdone/internal/service"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const (
	userIDKey      contextKey = "userID"
	jtiKey         contextKey = "jti"
	expirationTime contextKey = "expiresAt"
)

func AuthMiddleware(
	jwtService *service.JWTService,
	tokenBlackList *service.TokenBlacklist,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")

			if auth == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(auth, " ", 2)

			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			userID, jti, expiresAt, err := jwtService.Validate(token)

			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			revoked, err := tokenBlackList.IsRevoked(
				r.Context(),
				jti,
			)

			if err != nil {
				http.Error(
					w,
					"Internal Server Error",
					http.StatusInternalServerError,
				)
				return
			}

			if revoked {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userIDKey,
				userID,
			)

			ctx = context.WithValue(
				ctx,
				jtiKey,
				jti,
			)

			ctx = context.WithValue(
				ctx,
				expirationTime,
				expiresAt,
			)

			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func UserID(r *http.Request) (uuid.UUID, bool) {
	userID, ok := r.Context().Value(userIDKey).(uuid.UUID)

	return userID, ok
}

func JTI(r *http.Request) (string, bool) {
	jti, ok := r.Context().Value(jtiKey).(string)
	return jti, ok
}

func ExpiresAt(r *http.Request) (time.Time, bool) {
	expiresAt, ok := r.Context().Value(expirationTime).(time.Time)
	return expiresAt, ok
}
