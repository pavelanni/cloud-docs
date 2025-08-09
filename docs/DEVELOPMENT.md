# Cloud Docs development guide

This guide documents the complete development process, from initial setup to production deployment, including lessons learned and best practices.

## Development journey

### Phase 1: Project foundation (1-2 days)

**Objective**: Establish basic Go web server with modern tooling

**Key decisions made**:
- **Go 1.24**: Latest stable version with improved performance
- **Chi v5**: Modern HTTP router (replacing deprecated Gorilla Mux)
- **Standard project structure**: `cmd/`, `internal/`, `pkg/` organization
- **Configuration management**: Environment-based with defaults

**Implementation highlights**:
```go
// Modern Chi router setup
r := chi.NewRouter()
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
r.Use(middleware.Heartbeat("/ping"))
```

**Testing approach**:
- Unit tests for all configuration loading
- HTTP handler testing with httptest
- Build verification for all platforms

**Lessons learned**:
- Chi's middleware system is much cleaner than Gorilla
- Go's graceful shutdown patterns work excellently with Cloud Run
- Environment-based configuration scales better than config files

### Phase 2: Google Cloud Storage integration (2-3 days)

**Objective**: Implement secure, efficient file serving from GCS

**Technical challenges solved**:
1. **MIME type detection**: Custom logic for missing Content-Type headers
1. **Error handling**: Distinguishing between 404s and other errors
1. **Streaming**: Large file support without memory issues

**Key implementation**:
```go
// Efficient file streaming from GCS
func (c *Client) GetFile(ctx context.Context, objectPath string) (*FileInfo, error) {
    obj := c.client.Bucket(c.bucketName).Object(objectPath)
    reader, err := obj.NewReader(ctx)
    // Stream directly to HTTP response
    return &FileInfo{Content: reader, ContentType: contentType}, nil
}
```

**Testing strategy**:
- GCS emulator for unit tests
- Integration tests with real bucket
- Performance testing with various file sizes

**Lessons learned**:
- GCS client libraries handle authentication automatically
- Streaming is crucial for large files and memory efficiency
- Proper error handling improves debugging significantly

### Phase 3: Authentication middleware (2-3 days)

**Objective**: Implement secure, stateless token-based authentication

**Security design decisions**:
- **JWT-inspired format**: Easy to parse and validate
- **HMAC-SHA256**: Industry-standard signing algorithm
- **Multiple token sources**: Query params, headers, cookies for flexibility
- **Timing-safe comparison**: Prevent timing attacks

**Token format designed**:
```
{base64-payload}.{base64-signature}

Payload: {"id":"uuid","expires_at":"iso-date","issued_at":"iso-date"}
```

**Middleware implementation**:
```go
// Flexible token extraction
func extractToken(r *http.Request) string {
    // Priority: query param > header > cookie
    if token := r.URL.Query().Get("token"); token != "" {
        return token
    }
    if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
        return strings.TrimPrefix(auth, "Bearer ")
    }
    // ... cookie extraction
}
```

**Testing approach**:
- Comprehensive token generation/validation tests
- Middleware integration tests
- Security tests (tampering, expiration)

**Lessons learned**:
- Context propagation is essential for request tracing
- Multiple token sources improve integration flexibility
- Clear error messages aid debugging without information leakage

### Phase 4: Upload tool development (2-3 days)

**Objective**: Create production-ready CLI for document deployment

**Feature requirements identified**:
- Recursive directory upload with structure preservation
- Intelligent file filtering (exclude .git, .DS_Store, etc.)
- Progress reporting for large uploads
- Atomic operations with rollback capability

**CLI design patterns**:
```go
// Robust upload with progress tracking
type UploadStats struct {
    mu            sync.Mutex
    filesUploaded int
    filesSkipped  int
    totalBytes    int64
    errors        []error
}
```

**Filter implementation**:
```go
// Smart exclusion patterns
func shouldExclude(path string, patterns []string) bool {
    filename := filepath.Base(path)
    for _, pattern := range patterns {
        // Check both full path and filename
        if matched, _ := filepath.Match(pattern, path); matched {
            return true
        }
        if matched, _ := filepath.Match(pattern, filename); matched {
            return true
        }
    }
}
```

**Testing methodology**:
- Test with various directory structures
- Verify exclusion patterns work correctly
- Performance testing with large file sets

**Lessons learned**:
- Progress reporting dramatically improves user experience
- Default exclusion patterns prevent common mistakes
- Dry-run mode essential for testing and verification

### Phase 5: Iframe generation tool (1-2 days)

**Objective**: Simplify document embedding with automatic token handling

**Integration challenges**:
- URL encoding for tokens in query parameters
- HTML attribute handling and validation
- Token lifecycle management for embeds

**Implementation approach**:
```go
// Flexible iframe generation
func generateIframe(cfg *IframeConfig) (string, error) {
    documentURL, err := buildDocumentURL(cfg)
    if err != nil {
        return "", err
    }
    
    var attrs []string
    attrs = append(attrs, fmt.Sprintf(`src="%s"`, documentURL))
    // ... build all attributes
    
    return fmt.Sprintf("<iframe %s></iframe>\n", strings.Join(attrs, " ")), nil
}
```

**Token integration**:
```go
// Automatic token generation or use provided token
if iframeConfig.Token == "" {
    tokenManager := token.NewManager(cfg.TokenSecret)
    generatedToken, err := tokenManager.Generate(duration)
    iframeConfig.Token = generatedToken
}
```

**Testing focus**:
- URL encoding correctness
- HTML attribute validation
- Token integration end-to-end

**Lessons learned**:
- URL encoding is critical for token integrity
- Flexible attribute handling enables wide use cases
- Token auto-generation simplifies workflow significantly

### Phase 6: Deployment infrastructure (1-2 days)

**Objective**: Create production-ready containerization and deployment

**Container design decisions**:
- **Multi-stage build**: Minimize final image size
- **Alpine Linux**: Security and size benefits
- **Non-root execution**: Security hardening
- **Health checks**: Cloud Run integration

**Dockerfile optimization**:
```dockerfile
# Builder stage - full Go environment
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage - minimal environment  
FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup
COPY --from=builder /app/server ./
USER appuser
```

**Podman compatibility**:
- All scripts work with both Docker and Podman
- OCI format compatibility
- Local build and test workflows

**Deployment automation**:
- Cloud Build configuration
- Environment variable management
- IAM permission setup

**Lessons learned**:
- Multi-stage builds dramatically reduce image size
- Security hardening is straightforward with modern containers
- Podman compatibility requires minimal changes from Docker

### Phase 7: Production deployment and validation (1-2 days)

**Objective**: Deploy to production and validate all functionality

**Deployment challenges encountered**:
1. **Cloud Build substitution variables**: Required syntax adjustments
1. **IAM permissions**: Cloud Run authentication configuration
1. **Environment variable propagation**: Secrets management

**Solutions implemented**:
```yaml
# Corrected Cloud Build configuration
substitutions:
  _SERVICE_NAME: 'cloud-docs-server'
  _REGION: 'us-central1'
  _BUCKET_NAME: 'cloud-docs-storage-pavelanni2025'
```

**Validation test suite**:
- Basic endpoint functionality
- Authentication flow validation
- Document serving verification
- Security boundary testing
- Performance characteristics

**Production metrics observed**:
- Cold start: ~2-3 seconds
- Warm response: <100ms
- Authentication overhead: <5ms
- File serving: <500ms (including GCS fetch)

**Lessons learned**:
- Cloud Build substitution syntax is strict
- IAM configuration must match deployment expectations
- Comprehensive testing prevents production surprises
- Performance monitoring is essential from day one

## Best practices discovered

### Code organization
```
cmd/           # CLI tools and main applications
internal/      # Private application code
pkg/           # Public library code (reusable)
scripts/       # Build and deployment scripts
docs/          # Documentation
```

### Error handling patterns
```go
// Consistent error wrapping
if err != nil {
    return fmt.Errorf("failed to create client: %w", err)
}

// Security-conscious error messages
if strings.Contains(err.Error(), "file not found") {
    http.Error(w, "File not found", http.StatusNotFound)
    return
}
```

### Testing strategies
- **Unit tests**: All business logic
- **Integration tests**: External service interactions
- **End-to-end tests**: Complete workflows
- **Security tests**: Authentication and authorization
- **Performance tests**: Response times and scaling

### Configuration management
```go
// Environment with sensible defaults
func Load() *Config {
    return &Config{
        Port:        getEnv("PORT", "8080"),
        DocsPath:    getEnv("DOCS_PATH", "/docs"),
        LogLevel:    getEnv("LOG_LEVEL", "info"),
    }
}
```

## Development environment setup

### Required tools
```bash
# Core development
go install golang.org/dl/go1.24@latest  # Go toolchain
podman install                          # Container runtime
gcloud components install               # GCP CLI

# Optional but recommended
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Local development workflow
```bash
# 1. Build and test
go mod tidy
go test ./...
go build ./...

# 2. Local container testing
./scripts/build-local.sh test-project latest
podman run -p 8080:8080 -e TOKEN_SECRET=dev-secret cloud-docs-server:latest

# 3. Integration testing with GCS
export BUCKET_NAME=test-bucket
export TOKEN_SECRET=test-secret
go run ./cmd/server

# 4. CLI tool testing
./bin/upload -source test-docs -bucket test-bucket -dry-run
./bin/token -generate -expires 1h
./bin/iframe -document /test.html -base-url http://localhost:8080
```

### Code quality standards
- **Formatting**: `go fmt ./...`
- **Imports**: `goimports -w .`
- **Linting**: `golangci-lint run`
- **Testing**: `go test -race -cover ./...`
- **Security**: `gosec ./...`

## Debugging guide

### Common issues and solutions

#### "Token validation failed"
```bash
# Check token secret consistency
echo $TOKEN_SECRET | base64 -d | wc -c  # Should be 32+ chars
./bin/token -validate "your-token-here"
```

#### "File not found" errors
```bash
# Verify bucket contents
gsutil ls -r gs://your-bucket-name/
# Check file paths match exactly
curl -v "https://service-url/docs/exact/path?token=valid-token"
```

#### Container won't start
```bash
# Check container logs
podman logs container-name
# Verify environment variables
podman exec -it container-name env
```

#### GCS permission errors
```bash
# Check service account permissions
gcloud projects get-iam-policy project-id
# Test authentication
gcloud auth application-default login
gsutil ls gs://bucket-name
```

### Performance troubleshooting
```bash
# Check Cloud Run metrics
gcloud run services describe service-name --region=region

# Test response times
time curl -s "https://service-url/health"

# Monitor container resource usage
podman stats container-name
```

## Maintenance and updates

### Regular maintenance tasks
- **Dependency updates**: Monthly Go module updates
- **Security patches**: Container base image updates
- **Token rotation**: Periodic secret rotation
- **Metric review**: Monthly performance analysis

### Upgrade procedures
1. **Local testing**: All changes tested locally first
1. **Staging deployment**: Test environment validation
1. **Gradual rollout**: Cloud Run traffic splitting
1. **Monitoring**: Watch metrics during deployment
1. **Rollback plan**: Previous version ready if needed

### Monitoring dashboards
- **Request metrics**: QPS, latency, error rates
- **Resource usage**: CPU, memory, concurrent requests
- **Business metrics**: Document access patterns, token usage
- **Security metrics**: Failed authentication attempts

This development guide captures the complete journey from conception to production deployment, including all the challenges overcome and solutions discovered along the way.