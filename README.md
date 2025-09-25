# Cloud Docs

A Go-based serverless application for hosting private documents on Google Cloud Platform with token-based access control.

## Overview

Cloud Docs provides a secure, serverless solution for hosting HTML documents and static assets on
Google Cloud Platform. The system uses token-based authentication to keep documents private and hidden
from search engines, while providing an easy way to share documents with authorized users.

## Features

- **Token-based access control**: Documents are protected by URL tokens to prevent unauthorized access
- **Serverless deployment**: Runs on Google Cloud Run for automatic scaling and cost efficiency
- **Two-tier security model**: Protected documents require tokens, static assets (CSS/JS) are publicly accessible
- **Markdown to HTML conversion**: Built-in workflow for converting Markdown documentation to HTML
- **Mermaid diagram support**: Automatic conversion of Mermaid diagrams to images
- **Directory structure preservation**: Maintains original folder organization when uploading content
- **Iframe embedding**: Generate embeddable iframe elements with pre-authenticated URLs
- **CLI tools with pflag**: User-friendly command-line tools with both long and short flag options

## Architecture

The system consists of three main components:

### Web server
Go application deployed on Google Cloud Run that serves documents through a Chi router with token validation middleware. Provides two routes:
- `/docs/*` - Protected documents requiring token authentication
- `/docs/static/*` - Public static assets (CSS, JS, fonts) served without tokens

### Content conversion
Pandoc-based workflow that converts Markdown files to HTML with:
- Mermaid diagram processing using mermaid-cli
- Automatic token embedding for protected images and links
- CSS and static asset organization

### Token management
Tools for creating and managing access tokens that are embedded in document URLs and validated by the web server middleware.

## Technology stack

- **Language**: Go 1.24.6
- **HTTP router**: Chi v5
- **Cloud platform**: Google Cloud Platform
- **Storage**: Google Cloud Storage
- **Runtime**: Google Cloud Run
- **Authentication**: Token-based URL access control

## Quick start

### Prerequisites

- Go 1.24.6 or later
- Google Cloud SDK configured with appropriate permissions
- Google Cloud Project with Cloud Run and Cloud Storage APIs enabled

### Development

```bash
# Clone the repository
git clone <repository-url>
cd cloud-docs

# Install dependencies
go mod tidy

# Build the application
go build ./...

# Run tests
go test ./...

# Format code
go fmt ./...
```

### Deployment

Deploy to Google Cloud Run:

```bash
# Build and deploy
gcloud run deploy cloud-docs --source .
```

## Complete workflow

### 1. Generate long-lived access token

```bash
# Generate a token valid for 1 year
TOKEN=$(./cmd/token/token -g -e 8760h)
echo "Generated token: $TOKEN"
```

### 2. Convert Markdown to HTML

Convert your Markdown documentation to HTML with embedded tokens:

```bash
# Convert courses directory to HTML with token and static assets
./scripts/convert_all.sh \
  -s courses/ \
  -d output/ \
  -t "$TOKEN" \
  -c /path/to/static/assets

# Options:
#   -s, --src      Source directory containing Markdown files
#   -d, --dest     Destination directory for generated HTML
#   -t, --token    Authentication token for protected assets
#   -c, --css-dir  Directory containing static assets (optional)
```

This creates the following structure:
```
output/
├── courses/          # Converted markdown content
│   └── minio-for-admins/
│       └── network/
│           ├── README.html
│           └── images/
└── static/           # CSS, JS, fonts (if provided)
    └── css/
        └── minio_docs.css
```

### 3. Upload to Google Cloud Storage

Upload the processed content using the gcloud CLI:

```bash
# Upload all content to GCS bucket
gcloud storage cp -r output/* gs://your-bucket-name/

# For CI/CD environments, you might want to set up authentication first:
# gcloud auth activate-service-account --key-file=/path/to/service-account.json
```

### 4. Generate iframe for embedding

Create iframe elements for embedding in your LMS or website:

```bash
# Generate iframe for specific document
./cmd/iframe/iframe \
  -d "/courses/minio-for-admins/network/README.html" \
  -t "$TOKEN" \
  -u "https://your-cloud-run-url.com" \
  -w "100%" \
  -h "600"

# Options:
#   -d, --document  Path to the document
#   -t, --token     Access token
#   -u, --base-url  Base URL of Cloud Docs server
#   -w, --width     iframe width (default: 100%)
#   -h, --height    iframe height (default: 600)
#   -o, --output    Output file (default: stdout)
```

## Individual tool usage

### Token tool usage

```bash
# Generate new token with custom expiration
./cmd/token/token -g -e 24h

# Validate existing token
./cmd/token/token -v "your-token-here"

# Show help
./cmd/token/token -h
```

### Iframe generator

```bash
# Generate iframe with custom attributes
./cmd/iframe/iframe \
  -d "/path/to/doc.html" \
  -t "token" \
  -u "https://server.com" \
  --sandbox "allow-scripts allow-same-origin" \
  --title "My Document" \
  --class "embedded-doc"
```

## License

This project is licensed under the MIT License.