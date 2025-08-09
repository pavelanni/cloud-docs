# Cloud Docs API reference

This document provides complete API reference for the Cloud Docs service and CLI tools.

## HTTP API

### Base URL
```
https://your-service-url.run.app
```

### Authentication
All document endpoints require a valid access token. Tokens can be provided via:

1. **Query parameter**: `?token=your-token`
1. **Authorization header**: `Authorization: Bearer your-token`  
1. **Cookie**: `access_token=your-token`

### Public endpoints

#### GET /health
Health check endpoint for monitoring and load balancers.

**Response**:
```json
{
  "status": "ok",
  "timestamp": "2025-08-09T19:35:46.982317Z"
}
```

**Status codes**:
- `200 OK`: Service is healthy

---

#### GET /ping
Minimal health check endpoint (returns single dot).

**Response**: `.`

**Status codes**:
- `200 OK`: Service is responding

---

#### GET /
Root endpoint returning service identification.

**Response**: `Cloud Docs Server`

**Status codes**:
- `200 OK`: Service is running

### Static asset endpoints (public)

#### GET /docs/static/{path}
Serve static assets (CSS, JavaScript, images) without authentication for easier HTML integration.

**Parameters**:
- `path`: Asset path within static directory (e.g., `main.css`, `app.js`, `images/logo.png`)

**Examples**:
```bash
# Serve CSS file (no token required)
GET /docs/static/main.css

# Serve JavaScript file (no token required)  
GET /docs/static/app.js

# Serve image (no token required)
GET /docs/static/images/logo.png
```

**Response headers**:
```
Content-Type: text/css (or appropriate MIME type)
Cache-Control: public, max-age=3600
X-Content-Type-Options: nosniff
```

**Status codes**:
- `200 OK`: Asset served successfully
- `404 Not Found`: Asset does not exist or directory listing attempted

### Protected endpoints (authentication required)

#### GET /docs/{path}
Serve documents from Google Cloud Storage with token authentication.

**Parameters**:
- `path`: Document path (e.g., `index.html`, `folder/doc.html`)
- `token`: Access token (required)

**Examples**:
```bash
# Serve root document
GET /docs/?token=eyJpZCI6...

# Serve specific document  
GET /docs/guide/installation.html?token=eyJpZCI6...

# Note: Static assets are automatically served from /docs/static/ without tokens
# HTML can reference them normally: <link rel="stylesheet" href="static/main.css">
```

**Response headers**:
```
Content-Type: text/html; charset=utf-8
Content-Length: 1234
```

**Status codes**:
- `200 OK`: Document served successfully
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: Document does not exist
- `500 Internal Server Error`: Server or storage error

**Error responses**:
```
401 Unauthorized: "Access token required"
401 Unauthorized: "Invalid or expired token"  
404 Not Found: "File not found"
```

## CLI tools

### Upload tool (`./bin/upload`)

Upload documents to Google Cloud Storage while preserving directory structure.

#### Usage
```bash
./bin/upload [flags]
```

#### Flags
- `-source string`: Source directory to upload (default: `.`)
- `-bucket string`: GCS bucket name (or use `BUCKET_NAME` env var)
- `-prefix string`: Prefix to add to all uploaded files
- `-exclude string`: Comma-separated exclusion patterns  
- `-dry-run`: Show what would be uploaded without uploading
- `-verbose`: Verbose output

#### Examples
```bash
# Basic upload
./bin/upload -source ./docs -bucket my-docs-bucket

# Upload with prefix and exclusions
./bin/upload -source ./site -bucket prod-bucket \
  -prefix v2.0 -exclude "*.bak,temp/*,*.log"

# Dry run to preview
./bin/upload -source ./docs -bucket test-bucket -dry-run -verbose
```

#### Default exclusions
- `.git/*`: Git repository files
- `.DS_Store`: macOS system files  
- `*.tmp`: Temporary files
- `*.log`: Log files

#### Output format
```
Uploading from ./docs to gs://my-bucket with prefix v1
Uploading: css/main.css -> v1/css/main.css
Uploading: index.html -> v1/index.html
Skipping: temp.log (excluded)

Upload complete:
  Files uploaded: 12
  Files skipped: 3
  Total bytes: 45672 (0.04 MB)
  Duration: 2.341s
```

#### Exit codes
- `0`: Success
- `1`: Error (invalid arguments, upload failures)

---

### Token tool (`./bin/token`)

Generate and validate access tokens for document authentication.

#### Usage
```bash
./bin/token [flags]
```

#### Flags
- `-generate`: Generate a new token
- `-validate string`: Validate an existing token
- `-expires string`: Token expiration duration (default: `24h`)

#### Token secret
Set via `TOKEN_SECRET` environment variable or server configuration.

#### Examples
```bash
# Generate token (24h expiration)
./bin/token -generate

# Generate with custom expiration
./bin/token -generate -expires 72h

# Generate short-lived token  
./bin/token -generate -expires 30m

# Validate token
./bin/token -validate "eyJpZCI6..."
```

#### Duration formats
- `24h`: 24 hours
- `30m`: 30 minutes  
- `1h30m`: 1 hour 30 minutes
- `168`: 168 hours (plain number = hours)

#### Output formats

**Generation**:
```
Generated token (expires in 24h0m0s):
eyJpZCI6IjY5NmJiM2I0LTUwZGEtNDcwZi05OGM2LTgzNDVlMDBlNTAyZCIsImV4cGlyZXNfYXQiOiIyMDI1LTA4LTEwVDE5OjM0OjEwLjg4NzEyNVoiLCJpc3N1ZWRfYXQiOiIyMDI1LTA4LTA5VDE5OjM0OjEwLjg4NzEyNVoifQ==.NrKRFAdehr67yKUr7wupesxoY2nq7eNstygAYeF9pD8=
```

**Validation**:
```
Token is valid:
  ID: 696bb3b4-50da-470f-98c6-8345e00e502d
  Issued: 2025-08-09T19:34:10Z
  Expires: 2025-08-10T19:34:10Z
  Time left: 23h59m45s
```

#### Exit codes
- `0`: Success (token generated or valid)
- `1`: Error (invalid token, expired, generation failed)

---

### Iframe tool (`./bin/iframe`)

Generate HTML iframe elements with authenticated document URLs.

#### Usage
```bash
./bin/iframe [flags]
```

#### Required flags
- `-document string`: Path to document (e.g., `/guide/intro.html`)

#### Optional flags
- `-base-url string`: Service base URL (default: `http://localhost:8080`)
- `-docs-path string`: Docs path prefix (default: from `DOCS_PATH` env or `/docs`)
- `-token string`: Existing token (if not provided, generates new one)
- `-token-expires string`: Token expiration if generating (default: `24h`)

#### Iframe attributes
- `-width string`: iframe width (default: `100%`)
- `-height string`: iframe height (default: `600`)
- `-frameborder string`: Frame border (default: `0`)
- `-scrolling string`: Scrolling behavior (default: `auto`)
- `-allowfullscreen`: Allow fullscreen
- `-sandbox string`: Sandbox restrictions
- `-title string`: iframe title attribute
- `-class string`: CSS class name
- `-id string`: HTML id attribute  
- `-attrs string`: Custom attributes as `key=value,key2=value2`

#### Output options
- `-output string`: Write to file (default: stdout)
- `-verbose`: Verbose logging to stderr

#### Examples
```bash
# Basic iframe
./bin/iframe -document "/guide/intro.html"

# Production iframe with custom attributes
./bin/iframe -document "/api/reference.html" \
  -base-url "https://docs.myapp.com" \
  -width "800" -height "400" \
  -title "API Reference" \
  -class "api-docs" \
  -sandbox "allow-scripts allow-same-origin"

# With existing token  
./bin/iframe -document "/tutorial.html" \
  -token "eyJpZCI6..." \
  -output "embed.html"

# Custom attributes
./bin/iframe -document "/demo.html" \
  -attrs "data-testid=demo-iframe,loading=lazy"
```

#### Output format
```html
<iframe src="https://docs.myapp.com/docs/guide/intro.html?token=eyJpZCI6..." 
        width="800" height="400" frameborder="0" scrolling="auto" 
        title="API Reference" class="api-docs" 
        sandbox="allow-scripts allow-same-origin"></iframe>
```

#### Exit codes
- `0`: Success
- `1`: Error (invalid arguments, token generation failed)

## Token format specification

### Structure
```
{base64-encoded-payload}.{base64-encoded-signature}
```

### Payload format
```json
{
  "id": "696bb3b4-50da-470f-98c6-8345e00e502d",
  "expires_at": "2025-08-10T19:34:10.887125Z", 
  "issued_at": "2025-08-09T19:34:10.887125Z"
}
```

### Fields
- `id`: UUID v4 for request tracking
- `expires_at`: ISO 8601 timestamp (UTC) when token expires
- `issued_at`: ISO 8601 timestamp (UTC) when token was created

### Signature algorithm
- **Algorithm**: HMAC-SHA256
- **Key**: `TOKEN_SECRET` environment variable
- **Input**: Base64-encoded payload
- **Output**: Base64-encoded signature

### Validation rules
1. **Format**: Must have exactly one dot separator
1. **Payload**: Must be valid base64 and valid JSON
1. **Signature**: Must match HMAC-SHA256 of payload
1. **Expiration**: Current time must be before `expires_at`
1. **Fields**: All required fields must be present and valid

### Security considerations
- **Secret rotation**: Change `TOKEN_SECRET` to invalidate all tokens
- **Timing attacks**: Signature comparison uses constant-time algorithm
- **Token leakage**: Tokens in URLs may appear in logs/referrer headers
- **Expiration**: Short-lived tokens recommended for security

## Environment variables

### Server configuration
- `PORT`: HTTP server port (default: `8080`)
- `BUCKET_NAME`: Google Cloud Storage bucket name (required)
- `TOKEN_SECRET`: HMAC signing secret (required, base64-encoded recommended)
- `DOCS_PATH`: URL path prefix for documents (default: `/docs`)
- `LOG_LEVEL`: Logging level - `debug`, `info`, `warn`, `error` (default: `info`)

### CLI tool configuration
- `BUCKET_NAME`: Default bucket for upload tool
- `TOKEN_SECRET`: Secret for token generation/validation
- `DOCS_PATH`: Default docs path for iframe tool

### Google Cloud configuration
- `GOOGLE_APPLICATION_CREDENTIALS`: Service account key file path
- `GOOGLE_CLOUD_PROJECT`: GCP project ID (usually auto-detected)

## HTTP status codes

### Success codes
- `200 OK`: Request successful, content returned
- `204 No Content`: Request successful, no content

### Client error codes  
- `400 Bad Request`: Invalid request format or parameters
- `401 Unauthorized`: Missing, invalid, or expired authentication
- `404 Not Found`: Requested resource does not exist
- `405 Method Not Allowed`: HTTP method not supported for endpoint

### Server error codes
- `500 Internal Server Error`: Server error or storage system error
- `502 Bad Gateway`: Upstream service error (GCS unavailable)
- `503 Service Unavailable`: Service temporarily unavailable

## Rate limits and quotas

### Cloud Run limits
- **Concurrent requests**: 1000 per service (configurable)
- **Request timeout**: 15 minutes maximum
- **Instance memory**: 512MB-8GB per instance  
- **Instance CPU**: 0.5-4 vCPUs per instance

### Google Cloud Storage limits
- **Bandwidth**: 50 Gbps egress per bucket
- **Request rate**: 5000 requests per second per bucket
- **Object size**: 5TB maximum per object

### Token limits
- **Token size**: ~200-400 bytes typical
- **Expiration**: 1 minute to 365 days
- **Concurrent tokens**: No limit (stateless validation)

## Error handling

### Client errors
All client errors return plain text error messages:
```
Access token required
Invalid or expired token  
File not found
```

### Server errors
Server errors return generic messages to avoid information leakage:
```
Internal server error
```

Detailed error information is logged server-side for debugging.

### Logging format
```json
{
  "timestamp": "2025-08-09T19:34:10.887125Z",
  "level": "ERROR", 
  "message": "Failed to fetch file from storage",
  "path": "/docs/example.html",
  "error": "storage: object doesn't exist"
}
```