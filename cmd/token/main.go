package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pavelanni/cloud-docs/internal/config"
	"github.com/pavelanni/cloud-docs/pkg/token"
	"github.com/spf13/pflag"
)

func main() {
	var (
		generate = pflag.BoolP("generate", "g", false, "Generate a new token")
		validate = pflag.StringP("validate", "v", "", "Validate a token")
		expires  = pflag.StringP("expires", "e", "24h", "Token expiration duration (e.g., 24h, 7d, 168h)")
		help     = pflag.BoolP("help", "h", false, "Show help")
	)
	pflag.Parse()

	if *help {
		pflag.Usage()
		return
	}

	cfg := config.Load()
	tokenManager := token.NewManager(cfg.TokenSecret)

	if *generate {
		duration, err := token.ParseDuration(*expires)
		if err != nil {
			log.Fatalf("Invalid duration: %v", err)
		}

		tokenString, err := tokenManager.Generate(duration)
		if err != nil {
			log.Fatalf("Failed to generate token: %v", err)
		}

		fmt.Printf("Generated token (expires in %v):\n%s\n", duration, tokenString)
		return
	}

	if *validate != "" {
		validToken, err := tokenManager.Validate(*validate)
		if err != nil {
			fmt.Printf("Token validation failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Token is valid:\n")
		fmt.Printf("  ID: %s\n", validToken.ID)
		fmt.Printf("  Issued: %s\n", validToken.IssuedAt.Format(time.RFC3339))
		fmt.Printf("  Expires: %s\n", validToken.ExpiresAt.Format(time.RFC3339))
		fmt.Printf("  Time left: %v\n", time.Until(validToken.ExpiresAt).Round(time.Second))
		return
	}

	pflag.Usage()
}