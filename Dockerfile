# Multi-stage build for Cloud Run deployment
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Build the CLI tools
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o upload ./cmd/upload
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o token ./cmd/token
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o iframe ./cmd/iframe

# Final stage - minimal runtime image
FROM alpine:latest

# Install CA certificates and wget for health check
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/server /app/upload /app/token /app/iframe ./

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the server
CMD ["./server"]