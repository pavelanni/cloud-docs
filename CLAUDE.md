# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project overview

Cloud Docs is a Go-based serverless application for hosting documents on Google Cloud Platform. The system provides token-based access control for private document hosting and includes utilities for uploading content and generating embeddable iframe elements.

## Architecture

The project consists of three main components:

1. **Web server**: Go application deployed on Google Cloud Run that serves documents with token-based middleware authentication
1. **Upload tool**: Command-line utility to upload HTML/CSS/JS files to Google Cloud Storage while preserving directory structure
1. **Token management**: Tools for creating and managing access tokens for document URLs

## Technology stack

- **Language**: Go 1.24.6
- **HTTP router**: Chi v5 (lightweight, fast HTTP router with middleware support)
- **Cloud platform**: Google Cloud Platform
- **Storage**: Google Cloud Storage
- **Runtime**: Google Cloud Run (serverless)
- **Authentication**: Token-based URL access control

## Development commands

Since this is an early-stage project, common Go development commands apply:

```bash
# Initialize and tidy dependencies
go mod tidy

# Build the application
go build ./...

# Run tests
go test ./...

# Run a specific test
go test -run TestName ./path/to/package

# Format code
go fmt ./...

# Vet code
go vet ./...
```

## Project structure (planned)

Based on the requirements, the codebase will likely include:

- Web server with middleware for token validation
- Google Cloud Storage integration for file serving
- Command-line tools for file uploads and token generation
- Configuration for Google Cloud Run deployment