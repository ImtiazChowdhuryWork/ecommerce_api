package middleware

import (
	"context"
	"net/http"
	"strings"

	"ecommerce-api/internal/config"
	"ecommerce-api/internal/models"
	"ecommerce-api/internal/utils"
)

type contextKey string

const ClaimsKey contextKey = "claims"

type Auth struct {
	cfg *config.Config
}

func NewAuth(cfg *config.Config) *Auth {
	return &Auth{cfg: cfg}
}

// Authenticate validates the Bearer token and stores claims in context.
func (a *Auth) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			utils.WriteError(w, http.StatusUnauthorized, "authorization header required")
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.WriteError(w, http.StatusUnauthorized, "authorization format: Bearer <token>")
			return
		}

		claims, err := utils.ValidateToken(parts[1], a.cfg.JWTSecret)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		if claims.Type != "access" {
			utils.WriteError(w, http.StatusUnauthorized, "access token required")
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin must run after Authenticate.
func (a *Auth) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r)
		if claims == nil || claims.Role != models.RoleAdmin {
			utils.WriteError(w, http.StatusForbidden, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetClaims(r *http.Request) *utils.Claims {
	claims, _ := r.Context().Value(ClaimsKey).(*utils.Claims)
	return claims
}
