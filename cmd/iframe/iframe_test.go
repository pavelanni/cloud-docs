package main

import (
	"strings"
	"testing"
)

func TestParseCustomAttrs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "single attribute",
			input: "data-testid=iframe",
			expected: map[string]string{
				"data-testid": "iframe",
			},
		},
		{
			name:  "multiple attributes",
			input: "data-testid=iframe,loading=lazy,style=border:none",
			expected: map[string]string{
				"data-testid": "iframe",
				"loading":     "lazy",
				"style":       "border:none",
			},
		},
		{
			name:  "attributes with spaces",
			input: " data-testid = iframe , loading = lazy ",
			expected: map[string]string{
				"data-testid": "iframe",
				"loading":     "lazy",
			},
		},
		{
			name:     "invalid format",
			input:    "invalid,also-invalid",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCustomAttrs(tt.input)
			
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d attributes, got %d", len(tt.expected), len(result))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, exists := result[key]; !exists {
					t.Errorf("Expected key %q not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("Expected %q = %q, got %q", key, expectedValue, actualValue)
				}
			}
		})
	}
}

func TestBuildDocumentURL(t *testing.T) {
	tests := []struct {
		name        string
		config      *IframeConfig
		expectedURL string
		expectError bool
	}{
		{
			name: "basic URL",
			config: &IframeConfig{
				BaseURL:      "http://localhost:8080",
				DocsPath:     "/docs",
				DocumentPath: "/test/index.html",
				Token:        "test-token",
			},
			expectedURL: "http://localhost:8080/docs/test/index.html?token=test-token",
			expectError: false,
		},
		{
			name: "HTTPS URL with custom port",
			config: &IframeConfig{
				BaseURL:      "https://my-docs.example.com:8443",
				DocsPath:     "/documents",
				DocumentPath: "/folder/page.html",
				Token:        "secure-token",
			},
			expectedURL: "https://my-docs.example.com:8443/documents/folder/page.html?token=secure-token",
			expectError: false,
		},
		{
			name: "document path without leading slash",
			config: &IframeConfig{
				BaseURL:      "http://localhost:8080",
				DocsPath:     "/docs",
				DocumentPath: "test/index.html",
				Token:        "test-token",
			},
			expectedURL: "http://localhost:8080/docs/test/index.html?token=test-token",
			expectError: false,
		},
		{
			name: "relative path becomes absolute",
			config: &IframeConfig{
				BaseURL:      "relative-path",
				DocsPath:     "/docs",
				DocumentPath: "/test/index.html",
				Token:        "test-token",
			},
			expectedURL: "/docs/test/index.html?token=test-token",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := buildDocumentURL(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expectedURL {
				t.Errorf("Expected URL %q, got %q", tt.expectedURL, result)
			}
		})
	}
}

func TestGenerateIframe(t *testing.T) {
	tests := []struct {
		name           string
		config         *IframeConfig
		expectedAttrs  []string
		unexpectedAttr string
	}{
		{
			name: "basic iframe",
			config: &IframeConfig{
				BaseURL:      "http://localhost:8080",
				DocsPath:     "/docs",
				DocumentPath: "/test/index.html",
				Token:        "test-token",
				Width:        "100%",
				Height:       "600",
				Frameborder:  "0",
				Scrolling:    "auto",
			},
			expectedAttrs: []string{
				`src="http://localhost:8080/docs/test/index.html?token=test-token"`,
				`width="100%"`,
				`height="600"`,
				`frameborder="0"`,
				`scrolling="auto"`,
			},
		},
		{
			name: "iframe with all attributes",
			config: &IframeConfig{
				BaseURL:         "https://example.com",
				DocsPath:        "/docs",
				DocumentPath:    "/page.html",
				Token:           "token123",
				Width:           "800",
				Height:          "400",
				Frameborder:     "0",
				Scrolling:       "no",
				Allowfullscreen: true,
				Sandbox:         "allow-scripts allow-same-origin",
				Title:           "Test Document",
				Class:           "doc-frame",
				ID:              "main-doc",
				CustomAttrs: map[string]string{
					"data-testid": "iframe",
					"loading":     "lazy",
				},
			},
			expectedAttrs: []string{
				`src="https://example.com/docs/page.html?token=token123"`,
				`width="800"`,
				`height="400"`,
				`frameborder="0"`,
				`scrolling="no"`,
				`allowfullscreen`,
				`sandbox="allow-scripts allow-same-origin"`,
				`title="Test Document"`,
				`class="doc-frame"`,
				`id="main-doc"`,
				`data-testid="iframe"`,
				`loading="lazy"`,
			},
		},
		{
			name: "minimal iframe",
			config: &IframeConfig{
				BaseURL:      "http://localhost:8080",
				DocsPath:     "/docs",
				DocumentPath: "/minimal.html",
				Token:        "min-token",
			},
			expectedAttrs: []string{
				`src="http://localhost:8080/docs/minimal.html?token=min-token"`,
			},
			unexpectedAttr: `width=""`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generateIframe(tt.config)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if !strings.HasPrefix(result, "<iframe ") || !strings.HasSuffix(result, "></iframe>\n") {
				t.Error("Result should be a properly formatted iframe")
			}

			for _, expectedAttr := range tt.expectedAttrs {
				if !strings.Contains(result, expectedAttr) {
					t.Errorf("Expected attribute %q not found in result: %s", expectedAttr, result)
				}
			}

			if tt.unexpectedAttr != "" && strings.Contains(result, tt.unexpectedAttr) {
				t.Errorf("Unexpected attribute %q found in result: %s", tt.unexpectedAttr, result)
			}
		})
	}
}