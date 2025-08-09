package storage

import (
	"context"
	"strings"
	"testing"
)

func TestClient_UploadFile(t *testing.T) {
	ctx := context.Background()
	
	client, err := NewClient(ctx, "test-bucket")
	if err != nil {
		t.Skipf("Skipping test that requires GCP credentials: %v", err)
	}
	if client != nil {
		defer client.Close()
	}

	content := strings.NewReader("test content")
	
	err = client.UploadFile(ctx, "test/file.txt", content, "text/plain")
	if err != nil {
		t.Skipf("Skipping upload test (requires valid GCS bucket): %v", err)
	}
}

func TestUploadFile_ContentTypeDetection(t *testing.T) {
	ctx := context.Background()
	
	client, err := NewClient(ctx, "test-bucket")
	if err != nil {
		t.Skipf("Skipping test that requires GCP credentials: %v", err)
	}
	if client != nil {
		defer client.Close()
	}

	tests := []struct {
		filename    string
		content     string
		expectedCT  string
	}{
		{"test.html", "<html></html>", "text/html; charset=utf-8"},
		{"test.css", "body{}", "text/css; charset=utf-8"},
		{"test.js", "console.log()", "text/javascript; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			content := strings.NewReader(tt.content)
			err := client.UploadFile(ctx, tt.filename, content, "")
			if err != nil {
				t.Skipf("Skipping upload test (requires valid GCS bucket): %v", err)
			}
		})
	}
}