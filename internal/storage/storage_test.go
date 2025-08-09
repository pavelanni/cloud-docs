package storage

import (
	"context"
	"strings"
	"testing"
)

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		filename    string
		expected    string
	}{
		{"index.html", "text/html; charset=utf-8"},
		{"style.css", "text/css; charset=utf-8"},
		{"script.js", "text/javascript; charset=utf-8"},
		{"data.json", "application/json"},
		{"image.png", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"logo.svg", "image/svg+xml"},
		{"document.pdf", "application/pdf"},
		{"unknown.xyz", "chemical/x-xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := detectContentType(tt.filename)
			if result != tt.expected {
				t.Errorf("detectContentType(%s) = %s, want %s", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	ctx := context.Background()

	t.Run("empty bucket name", func(t *testing.T) {
		_, err := NewClient(ctx, "")
		if err == nil {
			t.Error("expected error for empty bucket name")
		}
		if !strings.Contains(err.Error(), "bucket name cannot be empty") {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("valid bucket name", func(t *testing.T) {
		client, err := NewClient(ctx, "test-bucket")
		if err != nil {
			t.Skipf("Skipping test that requires GCP credentials: %v", err)
		}
		if client == nil {
			t.Error("expected non-nil client")
		}
		if client != nil {
			client.Close()
		}
	})
}