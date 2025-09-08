package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/john221wick/golang-backend/learn/internal/auth"
)

type contextKey string

const AuthDataKey contextKey = "authData"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			// No Bearer prefix, try using the header as-is
			token = authHeader
		}

		authData, err := auth.ValidateToken(token)
		if err != nil {
			http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), AuthDataKey, authData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetAuthData(r *http.Request) *auth.AuthData {
	if authData, ok := r.Context().Value(AuthDataKey).(*auth.AuthData); ok {
		return authData
	}
	return nil
}
