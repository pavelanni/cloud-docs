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

func fileHandler(storageClient *storage.Client, docsPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, docsPath+"/")
		if path == "" {
			path = "index.html"
		}
		
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
		
		w.Header().Set("Content-Type", fileInfo.ContentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))
		
		if _, err := io.Copy(w, fileInfo.Content); err != nil {
			log.Printf("Error streaming file %s: %v", path, err)
		}
	}
}