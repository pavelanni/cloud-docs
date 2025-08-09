package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/pavelanni/cloud-docs/pkg/token"
)

type contextKey string

const TokenContextKey contextKey = "token"

func TokenMiddleware(tokenManager *token.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				http.Error(w, "Access token required", http.StatusUnauthorized)
				return
			}

			validToken, err := tokenManager.Validate(tokenString)
			if err != nil {
				// Don't log token details to avoid exposing tokens in Cloud Run logs
				log.Printf("Token validation failed for request to %s", r.URL.Path)
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), TokenContextKey, validToken)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	tokenString := r.URL.Query().Get("token")
	if tokenString != "" {
		return tokenString
	}

	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	if cookie, err := r.Cookie("access_token"); err == nil {
		return cookie.Value
	}

	return ""
}

func GetTokenFromContext(ctx context.Context) *token.Token {
	if token, ok := ctx.Value(TokenContextKey).(*token.Token); ok {
		return token
	}
	return nil
}