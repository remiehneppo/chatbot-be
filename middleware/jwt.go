package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/tieubaoca/chatbot-be/utils"
)

type JsonResponse struct {
	Error string `json:"error"`
}

type contextKey string

const (
	userContextKey  contextKey = "user"
	adminContextKey contextKey = "admin"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(JsonResponse{Error: "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(JsonResponse{Error: "Authorization header format must be Bearer {token}"})
			return
		}

		claims, err := utils.ParseUserToken(parts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			ctx := r.Context()
			ctx = context.WithValue(ctx, userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

	})
}

func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(JsonResponse{Error: "Authorization header is required"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(JsonResponse{Error: "Authorization header format must be Bearer {token}"})
			return
		}

		claims, err := utils.ParseAdminToken(parts[1])
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(JsonResponse{Error: "Invalid admin token"})
			return
		}

		if claims.Role != "admin" {
			ctx := r.Context()
			ctx = context.WithValue(ctx, adminContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}
	})
}
