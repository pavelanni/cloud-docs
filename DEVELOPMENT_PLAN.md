# Cloud Docs development plan

## Phase 1: Project setup and basic web server
**Duration: 1-2 days**

### Development tasks:
1. Create basic project structure:
   ```
   cmd/
     server/          # Main web server
     upload/          # Upload tool
     token/           # Token management tool
   internal/
     auth/            # Authentication middleware
     storage/         # GCS integration
     config/          # Configuration management
   pkg/
     token/           # Token utilities
   ```

1. Initialize Go modules and dependencies:
   ```bash
   go mod tidy
   go get cloud.google.com/go/storage
   go get github.com/go-chi/chi/v5
   go get github.com/go-chi/chi/v5/middleware
   go get github.com/google/uuid
   ```

1. Create basic HTTP server with health check endpoint
1. Add configuration management (environment variables)
1. Implement basic logging

### Testing:
- Unit tests for configuration loading
- Integration test for basic server startup
- Local server testing with curl

## Phase 2: Google Cloud Storage integration
**Duration: 2-3 days**

### Development tasks:
1. Create GCS client wrapper in `internal/storage/`
1. Implement file serving from GCS buckets
1. Handle MIME types for HTML, CSS, JS files
1. Add proper error handling for missing files
1. Implement directory structure preservation

### Testing:
- Unit tests with GCS emulator or mocks
- Integration tests with test bucket
- Test various file types (HTML, CSS, JS, images)

## Phase 3: Token-based authentication middleware
**Duration: 2-3 days**

### Development tasks:
1. Design token format and validation logic
1. Create middleware to extract and validate tokens from URLs
1. Implement token generation utilities
1. Add token expiration and refresh mechanisms
1. Create secure token storage/management

### Testing:
- Unit tests for token generation and validation
- Middleware integration tests
- Security tests for token tampering

## Phase 4: Upload tool development
**Duration: 2-3 days**

### Development tasks:
1. Create CLI tool in `cmd/upload/`
1. Implement recursive directory upload
1. Preserve directory structure in GCS
1. Add progress reporting and error handling
1. Support for file filtering and exclusions

### Testing:
- Unit tests for file processing logic
- Integration tests with test directories
- Test large file uploads and error scenarios

## Phase 5: iframe generation tool
**Duration: 1-2 days**

### Development tasks:
1. Create CLI tool in `cmd/iframe/` 
1. Generate iframe HTML with proper token URLs
1. Support customizable iframe attributes
1. Add URL validation and formatting

### Testing:
- Unit tests for iframe generation
- Integration tests with various document paths
- Validate generated HTML output

## Phase 6: GCP deployment setup
**Duration: 1-2 days**

### GCP setup tasks:
1. **Create GCP project:**
   ```bash
   gcloud projects create cloud-docs-[unique-id]
   gcloud config set project cloud-docs-[unique-id]
   ```

1. **Enable required APIs:**
   ```bash
   gcloud services enable storage.googleapis.com
   gcloud services enable run.googleapis.com
   gcloud services enable cloudbuild.googleapis.com
   ```

1. **Create Cloud Storage bucket:**
   ```bash
   gsutil mb gs://cloud-docs-storage-[unique-id]
   gsutil lifecycle set lifecycle.json gs://cloud-docs-storage-[unique-id]
   ```

1. **Set up IAM permissions:**
   ```bash
   gcloud projects add-iam-policy-binding cloud-docs-[unique-id] \
     --member="serviceAccount:cloud-run-service@cloud-docs-[unique-id].iam.gserviceaccount.com" \
     --role="roles/storage.objectViewer"
   ```

1. **Create Dockerfile for Cloud Run:**
   ```dockerfile
   FROM golang:1.24-alpine AS builder
   WORKDIR /app
   COPY . .
   RUN go mod download
   RUN go build -o server ./cmd/server

   FROM alpine:latest
   RUN apk --no-cache add ca-certificates
   WORKDIR /root/
   COPY --from=builder /app/server .
   EXPOSE 8080
   CMD ["./server"]
   ```

## Phase 7: Cloud Run deployment and testing
**Duration: 1-2 days**

### Deployment tasks:
1. **Build and deploy to Cloud Run:**
   ```bash
   gcloud builds submit --tag gcr.io/cloud-docs-[unique-id]/server
   gcloud run deploy cloud-docs-server \
     --image gcr.io/cloud-docs-[unique-id]/server \
     --platform managed \
     --region us-central1 \
     --set-env-vars BUCKET_NAME=cloud-docs-storage-[unique-id]
   ```

1. **Upload test documents:**
   ```bash
   ./upload-tool --bucket cloud-docs-storage-[unique-id] --source ./test-docs/
   ```

1. **Generate test tokens and URLs:**
   ```bash
   ./token-tool --generate --expires 24h
   ./iframe-tool --document /path/to/doc.html --token [generated-token]
   ```

### Testing:
- End-to-end testing with deployed service
- Load testing with multiple concurrent requests
- Security testing for token validation
- Performance testing for large documents

## Additional considerations

### Security:
- Implement rate limiting
- Add request logging and monitoring
- Use HTTPS only
- Validate all input parameters
- Implement proper CORS headers

### Monitoring and operations:
- Add Cloud Monitoring integration
- Implement structured logging
- Set up error reporting
- Create health check endpoints
- Add metrics for token usage

### Performance optimizations:
- Implement caching headers
- Add CDN integration (Cloud CDN)
- Optimize for Cold Start performance
- Implement connection pooling for GCS

This phased approach allows for incremental development, testing, and deployment while building toward the complete system outlined in the README.