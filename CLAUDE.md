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

## Complete workflow for content processing

The system now includes a comprehensive workflow for converting Markdown documentation to secure, token-protected HTML:

### 1. Content conversion workflow

```bash
# Generate long-lived token (1 year)
TOKEN=$(./cmd/token/token -g -e 8760h)

# Convert Markdown to HTML with embedded tokens
./scripts/convert_all.sh \
  -s courses/ \
  -d output/ \
  -t "$TOKEN" \
  -c /path/to/static/assets

# Upload processed content to GCS
./cmd/upload/upload -s output/ -b your-bucket-name

# Generate iframe for embedding
./cmd/iframe/iframe \
  -d "/courses/minio-for-admins/network/README.html" \
  -t "$TOKEN" \
  -u "https://your-cloud-run-url.com"
```

### 2. CLI tools (using pflag)

All CLI tools support both long and short flags:

**Token tool:**
- `-g, --generate`: Generate new token
- `-v, --validate`: Validate existing token  
- `-e, --expires`: Token expiration duration
- `-h, --help`: Show help

**Upload tool:**
- `-s, --source`: Source directory
- `-b, --bucket`: GCS bucket name
- `-p, --prefix`: Upload prefix
- `-e, --exclude`: Exclude patterns
- `-d, --dry-run`: Dry run mode
- `-v, --verbose`: Verbose output

**Iframe tool:**
- `-d, --document`: Document path
- `-t, --token`: Access token
- `-u, --base-url`: Server base URL
- `-w, --width`: iframe width
- `-o, --output`: Output file

### 3. Content conversion script

The `scripts/convert_all.sh` script handles:
- Markdown to HTML conversion using Pandoc
- Mermaid diagram processing with mermaid-cli
- Dynamic token embedding via Lua filter
- Static asset organization
- Directory structure preservation

Output structure:
```
output/
├── courses/          # Protected content (requires tokens)
└── static/           # Public assets (CSS, JS, fonts)
```

## Project structure

The codebase includes:

- **Web server** (`cmd/server/`): Chi-based HTTP server with token validation middleware
- **CLI tools** (`cmd/`): pflag-based command-line utilities for token, upload, and iframe operations
- **Internal packages** (`internal/`): Configuration, authentication, and storage abstractions
- **Token package** (`pkg/token/`): JWT-based token generation and validation
- **Conversion scripts** (`scripts/`): Pandoc-based Markdown to HTML conversion with Mermaid support
- **Deployment** (`scripts/deploy.sh`, `cloudbuild.yaml`): Google Cloud Run deployment configuration