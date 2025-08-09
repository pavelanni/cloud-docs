package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pavelanni/cloud-docs/pkg/token"
)

func TestTokenMiddleware(t *testing.T) {
	tokenManager := token.NewManager("test-secret")
	middleware := TokenMiddleware(tokenManager)

	validToken, err := tokenManager.Generate(24 * time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	expiredToken, err := tokenManager.Generate(-time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenFromCtx := GetTokenFromContext(r.Context())
		if tokenFromCtx == nil {
			t.Error("Expected token in context")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	tests := []struct {
		name           string
		setupRequest   func(*http.Request)
		expectedStatus int
	}{
		{
			name: "valid token in query param",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Set("token", validToken)
				r.URL.RawQuery = q.Encode()
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid token in authorization header",
			setupRequest: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer "+validToken)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid token in cookie",
			setupRequest: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: validToken,
				})
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no token",
			setupRequest:   func(r *http.Request) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "expired token",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Set("token", expiredToken)
				r.URL.RawQuery = q.Encode()
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			setupRequest: func(r *http.Request) {
				q := r.URL.Query()
				q.Set("token", "invalid-token")
				r.URL.RawQuery = q.Encode()
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			tt.setupRequest(req)

			rr := httptest.NewRecorder()
			middleware(handler).ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestExtractToken(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*http.Request)
		expected string
	}{
		{
			name: "from query param",
			setup: func(r *http.Request) {
				q := r.URL.Query()
				q.Set("token", "query-token")
				r.URL.RawQuery = q.Encode()
			},
			expected: "query-token",
		},
		{
			name: "from authorization header",
			setup: func(r *http.Request) {
				r.Header.Set("Authorization", "Bearer header-token")
			},
			expected: "header-token",
		},
		{
			name: "from cookie",
			setup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{
					Name:  "access_token",
					Value: "cookie-token",
				})
			},
			expected: "cookie-token",
		},
		{
			name: "priority: query over header",
			setup: func(r *http.Request) {
				q := r.URL.Query()
				q.Set("token", "query-token")
				r.URL.RawQuery = q.Encode()
				r.Header.Set("Authorization", "Bearer header-token")
			},
			expected: "query-token",
		},
		{
			name:     "no token",
			setup:    func(r *http.Request) {},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			tt.setup(req)

			result := extractToken(req)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}