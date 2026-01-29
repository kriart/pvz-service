package middleware

import (
	"context"
	"net/http"
	"strings"

	"pvz-service/internal/domain/user"
	"pvz-service/internal/usecase/ports"
)

type contextKey string

const userContextKey contextKey = "authUser"

func AuthMiddleware(tokenManager ports.TokenManager) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			usr, err := tokenManager.ParseToken(tokenStr)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userContextKey, usr)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool)
	for _, role := range allowedRoles {
		roleSet[role] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := r.Context().Value(userContextKey)
			if val == nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			u, ok := val.(*user.User)
			if !ok || !roleSet[u.Role] {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
