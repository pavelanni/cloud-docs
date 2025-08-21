package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/pavelanni/cloud-docs/internal/config"
	"github.com/pavelanni/cloud-docs/pkg/token"
	"github.com/spf13/pflag"
)

type IframeConfig struct {
	BaseURL         string
	DocsPath        string
	DocumentPath    string
	Token           string
	Width           string
	Height          string
	Frameborder     string
	Scrolling       string
	Allowfullscreen bool
	Sandbox         string
	Title           string
	Class           string
	ID              string
	CustomAttrs     map[string]string
}

func main() {
	var (
		baseURL         = pflag.StringP("base-url", "u", "", "Base URL of the Cloud Docs server (e.g., https://my-server.com)")
		docsPath        = pflag.String("docs-path", "", "Docs path prefix (default from DOCS_PATH env or /docs)")
		documentPath    = pflag.StringP("document", "d", "", "Path to the document (e.g., /folder/doc.html)")
		tokenString     = pflag.StringP("token", "t", "", "Access token (if not provided, will generate one)")
		tokenExpires    = pflag.String("token-expires", "24h", "Token expiration if generating new token")
		width           = pflag.StringP("width", "w", "100%", "iframe width")
		height          = pflag.String("height", "600", "iframe height")
		frameborder     = pflag.String("frameborder", "0", "iframe frameborder")
		scrolling       = pflag.String("scrolling", "auto", "iframe scrolling")
		allowfullscreen = pflag.Bool("allowfullscreen", false, "Allow fullscreen")
		sandbox         = pflag.String("sandbox", "", "Sandbox restrictions (e.g., 'allow-scripts allow-same-origin')")
		title           = pflag.String("title", "", "iframe title attribute")
		class           = pflag.String("class", "", "iframe CSS class")
		id              = pflag.String("id", "", "iframe ID attribute")
		customAttrs     = pflag.String("attrs", "", "Custom attributes as key=value,key2=value2")
		output          = pflag.StringP("output", "o", "", "Output file (default: stdout)")
		verbose         = pflag.BoolP("verbose", "v", false, "Verbose output")
		help            = pflag.BoolP("help", "h", false, "Show help")
	)
	pflag.Parse()

	if *help {
		pflag.Usage()
		return
	}

	if *documentPath == "" {
		pflag.Usage()
		log.Fatal("Document path is required")
	}

	cfg := config.Load()

	if *baseURL == "" {
		*baseURL = "http://localhost:" + cfg.Port
	}

	if *docsPath == "" {
		*docsPath = cfg.DocsPath
	}

	iframeConfig := &IframeConfig{
		BaseURL:         *baseURL,
		DocsPath:        *docsPath,
		DocumentPath:    *documentPath,
		Token:           *tokenString,
		Width:           *width,
		Height:          *height,
		Frameborder:     *frameborder,
		Scrolling:       *scrolling,
		Allowfullscreen: *allowfullscreen,
		Sandbox:         *sandbox,
		Title:           *title,
		Class:           *class,
		ID:              *id,
		CustomAttrs:     parseCustomAttrs(*customAttrs),
	}

	if iframeConfig.Token == "" {
		tokenManager := token.NewManager(cfg.TokenSecret)
		duration, err := token.ParseDuration(*tokenExpires)
		if err != nil {
			log.Fatalf("Invalid token expiration: %v", err)
		}

		generatedToken, err := tokenManager.Generate(duration)
		if err != nil {
			log.Fatalf("Failed to generate token: %v", err)
		}

		iframeConfig.Token = generatedToken

		if *verbose {
			fmt.Fprintf(os.Stderr, "Generated token (expires in %v): %s\n", duration, generatedToken)
		}
	}

	iframeHTML, err := generateIframe(iframeConfig)
	if err != nil {
		log.Fatalf("Failed to generate iframe: %v", err)
	}

	if *output != "" {
		file, err := os.Create(*output)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer file.Close()

		if _, err := file.WriteString(iframeHTML); err != nil {
			log.Fatalf("Failed to write to output file: %v", err)
		}

		if *verbose {
			fmt.Fprintf(os.Stderr, "Iframe HTML written to %s\n", *output)
		}
	} else {
		fmt.Print(iframeHTML)
	}
}

func parseCustomAttrs(attrs string) map[string]string {
	result := make(map[string]string)
	if attrs == "" {
		return result
	}

	pairs := strings.Split(attrs, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				result[key] = value
			}
		}
	}

	return result
}

func generateIframe(cfg *IframeConfig) (string, error) {
	documentURL, err := buildDocumentURL(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to build document URL: %w", err)
	}

	var attrs []string

	attrs = append(attrs, fmt.Sprintf(`src="%s"`, documentURL))
	
	// Security: Prevent referrer leakage
	attrs = append(attrs, `referrerpolicy="no-referrer"`)

	if cfg.Width != "" {
		attrs = append(attrs, fmt.Sprintf(`width="%s"`, cfg.Width))
	}

	if cfg.Height != "" {
		attrs = append(attrs, fmt.Sprintf(`height="%s"`, cfg.Height))
	}

	if cfg.Frameborder != "" {
		attrs = append(attrs, fmt.Sprintf(`frameborder="%s"`, cfg.Frameborder))
	}

	if cfg.Scrolling != "" {
		attrs = append(attrs, fmt.Sprintf(`scrolling="%s"`, cfg.Scrolling))
	}

	// Always allow clipboard access for copy functionality
	attrs = append(attrs, `allow="clipboard-write"`)

	if cfg.Allowfullscreen {
		attrs = append(attrs, "allowfullscreen")
	}

	if cfg.Sandbox != "" {
		attrs = append(attrs, fmt.Sprintf(`sandbox="%s"`, cfg.Sandbox))
	}

	if cfg.Title != "" {
		attrs = append(attrs, fmt.Sprintf(`title="%s"`, cfg.Title))
	}

	if cfg.Class != "" {
		attrs = append(attrs, fmt.Sprintf(`class="%s"`, cfg.Class))
	}

	if cfg.ID != "" {
		attrs = append(attrs, fmt.Sprintf(`id="%s"`, cfg.ID))
	}

	for key, value := range cfg.CustomAttrs {
		attrs = append(attrs, fmt.Sprintf(`%s="%s"`, key, value))
	}

	return fmt.Sprintf("<iframe %s></iframe>\n", strings.Join(attrs, " ")), nil
}

func buildDocumentURL(cfg *IframeConfig) (string, error) {
	baseURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}

	documentPath := cfg.DocumentPath
	if !strings.HasPrefix(documentPath, "/") {
		documentPath = "/" + documentPath
	}

	fullPath := cfg.DocsPath + documentPath

	documentURL := &url.URL{
		Scheme: baseURL.Scheme,
		Host:   baseURL.Host,
		Path:   fullPath,
	}

	query := documentURL.Query()
	query.Set("token", cfg.Token)
	documentURL.RawQuery = query.Encode()

	return documentURL.String(), nil
}