package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pavelanni/cloud-docs/internal/auth"
	"github.com/pavelanni/cloud-docs/internal/config"
	"github.com/pavelanni/cloud-docs/internal/storage"
	"github.com/pavelanni/cloud-docs/pkg/token"
)

func main() {
	cfg := config.Load()
	
	tokenManager := token.NewManager(cfg.TokenSecret)
	
	var storageClient *storage.Client
	if cfg.BucketName != "" {
		var err error
		storageClient, err = storage.NewClient(context.Background(), cfg.BucketName)
		if err != nil {
			log.Fatalf("Failed to create storage client: %v", err)
		}
		defer storageClient.Close()
		log.Printf("Connected to GCS bucket: %s", cfg.BucketName)
	} else {
		log.Println("No bucket configured, file serving disabled")
	}
	
	r := chi.NewRouter()
	
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Heartbeat("/ping"))
	
	r.Get("/health", healthHandler)
	r.Get("/", rootHandler)
	
	if storageClient != nil {
		// Serve static assets (CSS, JS, images) without token for easier HTML integration
		r.Route(cfg.DocsPath+"/static", func(r chi.Router) {
			r.Get("/*", staticFileHandler(storageClient, cfg.DocsPath+"/static"))
		})
		
		// Serve documents with token authentication (HTML and other content)
		r.Route(cfg.DocsPath, func(r chi.Router) {
			r.Use(auth.TokenMiddleware(tokenManager))
			r.Get("/*", fileHandler(storageClient, cfg.DocsPath))
		})
	}
	
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}
	
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server stopped")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Cloud Docs Server\n")
}

func staticFileHandler(storageClient *storage.Client, staticPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, staticPath+"/")
		log.Printf("Static file request: URL=%s, trimmed path=%s", r.URL.Path, path)

		// Only serve files from static/ directory - no index.html fallback
		if path == "" || strings.HasSuffix(path, "/") {
			log.Printf("Static file rejected: empty path or directory")
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Prepend "static/" to the path to ensure we're serving from static directory
		path = "static/" + path
		log.Printf("Looking for file in storage: %s", path)

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		fileInfo, err := storageClient.GetFile(ctx, path)
		if err != nil {
			log.Printf("Error getting static file %s: %v", path, err)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		defer fileInfo.Content.Close()
		log.Printf("Successfully serving static file: %s (size: %d bytes)", path, fileInfo.Size)

		// Security headers for static assets (but less restrictive than documents)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Type", fileInfo.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))

		// Public cache for static assets (since they don't contain sensitive data)
		w.Header().Set("Cache-Control", "public, max-age=3600")

		// Use io.CopyN to ensure we copy exactly the expected number of bytes
		n, err := io.CopyN(w, fileInfo.Content, fileInfo.Size)
		if err != nil && err != io.EOF {
			log.Printf("Error streaming static file %s: copied %d of %d bytes, error: %v", path, n, fileInfo.Size, err)
		} else {
			log.Printf("Successfully streamed static file %s: %d bytes", path, n)
		}
	}
}

func fileHandler(storageClient *storage.Client, docsPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, docsPath+"/")

		// Redirect static file requests to the static handler route
		if strings.HasPrefix(path, "static/") {
			http.Redirect(w, r, docsPath+"/"+path, http.StatusMovedPermanently)
			return
		}

		// Prevent directory listing - reject paths ending with / or empty paths without explicit file
		if path == "" {
			path = "index.html"
		} else if strings.HasSuffix(path, "/") {
			http.Error(w, "Directory listing not allowed", http.StatusForbidden)
			return
		}

		log.Printf("Document request: path=%s", path)

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		fileInfo, err := storageClient.GetFile(ctx, path)
		if err != nil {
			if strings.Contains(err.Error(), "file not found") {
				http.Error(w, "File not found", http.StatusNotFound)
				return
			}
			log.Printf("Error serving file %s: %v", path, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer fileInfo.Content.Close()
		log.Printf("Serving document %s (size: %d bytes, type: %s)", path, fileInfo.Size, fileInfo.ContentType)

		// Security headers to prevent indexing and improve security
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive, nosnippet")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "ALLOWALL")
		w.Header().Set("Referrer-Policy", "no-referrer")

		// Content-Type and caching based on file type
		w.Header().Set("Content-Type", fileInfo.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))

		// Cache policy based on content type
		if strings.Contains(fileInfo.ContentType, "text/html") {
			w.Header().Set("Cache-Control", "private, max-age=60")
		} else {
			// CSS, JS, images - longer cache but still private since auth required
			w.Header().Set("Cache-Control", "private, max-age=3600")
		}

		// Use io.CopyN to ensure we copy exactly the expected number of bytes
		n, err := io.CopyN(w, fileInfo.Content, fileInfo.Size)
		if err != nil && err != io.EOF {
			log.Printf("Error streaming file %s: copied %d of %d bytes, error: %v", path, n, fileInfo.Size, err)
		} else {
			log.Printf("Successfully streamed file %s: %d bytes", path, n)
		}
	}
}