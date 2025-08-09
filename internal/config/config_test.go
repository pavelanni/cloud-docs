package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Config
	}{
		{
			name:    "default values",
			envVars: map[string]string{},
			expected: Config{
				Port:        "8080",
				BucketName:  "",
				TokenSecret: "default-secret-change-in-production",
				LogLevel:    "info",
				DocsPath:    "/docs",
			},
		},
		{
			name: "custom values",
			envVars: map[string]string{
				"PORT":         "9000",
				"BUCKET_NAME":  "test-bucket",
				"TOKEN_SECRET": "test-secret",
				"LOG_LEVEL":    "debug",
				"DOCS_PATH":    "/documents",
			},
			expected: Config{
				Port:        "9000",
				BucketName:  "test-bucket",
				TokenSecret: "test-secret",
				LogLevel:    "debug",
				DocsPath:    "/documents",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			cfg := Load()

			if cfg.Port != tt.expected.Port {
				t.Errorf("Port = %v, want %v", cfg.Port, tt.expected.Port)
			}
			if cfg.BucketName != tt.expected.BucketName {
				t.Errorf("BucketName = %v, want %v", cfg.BucketName, tt.expected.BucketName)
			}
			if cfg.TokenSecret != tt.expected.TokenSecret {
				t.Errorf("TokenSecret = %v, want %v", cfg.TokenSecret, tt.expected.TokenSecret)
			}
			if cfg.LogLevel != tt.expected.LogLevel {
				t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, tt.expected.LogLevel)
			}
			if cfg.DocsPath != tt.expected.DocsPath {
				t.Errorf("DocsPath = %v, want %v", cfg.DocsPath, tt.expected.DocsPath)
			}
		})
	}
}