# Cloud Docs

A Go-based serverless application for hosting private documents on Google Cloud Platform with token-based access control.

## Overview

Cloud Docs provides a secure, serverless solution for hosting HTML documents and static assets on
Google Cloud Platform. The system uses token-based authentication to keep documents private and hidden
from search engines, while providing an easy way to share documents with authorized users.

## Features

- **Token-based access control**: Documents are protected by URL tokens to prevent unauthorized access
- **Serverless deployment**: Runs on Google Cloud Run for automatic scaling and cost efficiency
- **Static asset serving**: Serves HTML, CSS, JS, and other resources from Google Cloud Storage
- **Directory structure preservation**: Maintains original folder organization when uploading content
- **Iframe embedding**: Generate embeddable iframe elements with pre-authenticated URLs

## Architecture

The system consists of three main components:

### Web server
Go application deployed on Google Cloud Run that serves documents through a Chi router with token validation middleware.

### Upload tool
Command-line utility for uploading HTML, CSS, JS, and other files to Google Cloud Storage while preserving directory structure.

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

## Usage

### Uploading documents

Use the upload tool to upload HTML documents and assets to Cloud Storage:

```bash
# Upload a document directory
./upload-tool --source ./my-docs --bucket my-cloud-docs-bucket
```

### Generating access tokens

Create tokens for document access:

```bash
# Generate a new access token
./token-tool generate --document-path /my-document.html
```

### Embedding documents

Generate iframe elements for embedding documents:

```bash
# Create iframe with authenticated URL
./iframe-tool --document-path /my-document.html --token <access-token>
```

## License

This project is licensed under the MIT License.