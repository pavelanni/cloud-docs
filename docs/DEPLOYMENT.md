# Cloud Docs deployment guide

This guide covers deploying the Cloud Docs application to Google Cloud Run.

## Customizing for your organization

Before deploying, you'll need to customize several references to match your organization's naming conventions and repository structure.

### Required changes

#### 1. Go module path
**Current**: `github.com/pavelanni/cloud-docs`
**Change to**: Your organization's repository path

```bash
# Update the module path in go.mod
go mod edit -module github.com/yourcompany/cloud-docs

# Update all import statements in Go files
find . -name "*.go" -exec sed -i 's|github.com/pavelanni/cloud-docs|github.com/yourcompany/cloud-docs|g' {} +

# Tidy dependencies
go mod tidy
```

#### 2. Cloud Build configuration
**File**: `cloudbuild.yaml`
```yaml
# Change this line:
_BUCKET_NAME: 'cloud-docs-storage-pavelanni2025'
# To your organization's bucket name:
_BUCKET_NAME: 'yourcompany-cloud-docs-storage'
```

#### 3. Environment template
**File**: `.env.example`
```bash
# Change:
BUCKET_NAME=cloud-docs-storage-your-project-id
# To:
BUCKET_NAME=yourcompany-cloud-docs-storage
```

#### 4. Update example commands in this documentation
Replace all example project and bucket names:
- `my-cloud-docs-project` → `yourcompany-cloud-docs`
- `cloud-docs-storage-your-project` → `yourcompany-cloud-docs-storage`
- `cloud-docs-storage-pavelanni2025` → `yourcompany-cloud-docs-storage`

### Deployment naming recommendations

**Project naming**:
- Development: `yourcompany-cloud-docs-dev`
- Staging: `yourcompany-cloud-docs-staging` 
- Production: `yourcompany-cloud-docs-prod`

**Bucket naming**:
- Development: `yourcompany-cloud-docs-dev-storage`
- Staging: `yourcompany-cloud-docs-staging-storage`
- Production: `yourcompany-cloud-docs-prod-storage`

**Service naming**:
- Keep default: `cloud-docs-server` (configured in scripts)

## Prerequisites

- Google Cloud SDK (`gcloud`) installed and configured
- Container runtime (Docker or Podman) installed
- A Google Cloud Platform account with billing enabled
- Basic familiarity with GCP services

## Quick start

### 1. Set up GCP project

```bash
# Run the setup script (replace with your organization's naming)
./scripts/setup-gcp.sh yourcompany-cloud-docs-prod

# Follow the prompts and note the bucket name created
```

### 2. Upload your documents

```bash
# Upload your HTML/CSS/JS files (use your customized bucket name)
./bin/upload -source ./your-docs -bucket yourcompany-cloud-docs-prod-storage

# Or use a custom bucket name
./bin/upload -source ./your-docs -bucket your-custom-bucket-name
```

### 3. Deploy to Cloud Run

```bash
# Generate a secure token secret
TOKEN_SECRET=$(openssl rand -base64 32)

# Deploy the application (use your project and bucket names)
./scripts/deploy.sh yourcompany-cloud-docs-prod yourcompany-cloud-docs-prod-storage "$TOKEN_SECRET"
```

### 4. Test the deployment

```bash
# Get the service URL
SERVICE_URL=$(gcloud run services describe cloud-docs-server --region=us-central1 --format='value(status.url)')

# Test health endpoint
curl $SERVICE_URL/health

# Generate a token for testing
./bin/token -generate -expires 1h

# Test document access (replace TOKEN with generated token)
curl "$SERVICE_URL/docs/your-document.html?token=TOKEN"
```

## Manual deployment steps

### 1. Create GCP project and enable APIs

```bash
# Create project (use your organization's naming)
gcloud projects create yourcompany-cloud-docs-prod
gcloud config set project yourcompany-cloud-docs-prod

# Enable required APIs
gcloud services enable storage.googleapis.com
gcloud services enable run.googleapis.com
gcloud services enable cloudbuild.googleapis.com
```

### 2. Create Cloud Storage bucket

```bash
# Create bucket (use your organization's naming)
gsutil mb -l us-central1 gs://yourcompany-cloud-docs-prod-storage

# Upload documents
gsutil -m cp -r your-docs/* gs://yourcompany-cloud-docs-prod-storage/
```

### 3. Build and deploy with Cloud Build

```bash
# Submit build (use your bucket name and secure secret)
gcloud builds submit . \
    --config cloudbuild.yaml \
    --substitutions _BUCKET_NAME=yourcompany-cloud-docs-prod-storage,_TOKEN_SECRET=your-secure-secret
```

### 4. Configure environment variables (optional)

```bash
# Update service with additional environment variables
gcloud run services update cloud-docs-server \
    --region=us-central1 \
    --set-env-vars DOCS_PATH=/documents,LOG_LEVEL=info
```

## Local development and testing

### Build locally with Podman

```bash
# Build the image (use your project ID)
./scripts/build-local.sh yourcompany-cloud-docs-prod latest

# Run locally without GCS (basic endpoints only)
podman run -p 8080:8080 -e TOKEN_SECRET=dev-secret cloud-docs-server:latest

# Run with GCS integration (requires credentials)
podman run -p 8080:8080 \
    -e BUCKET_NAME=yourcompany-cloud-docs-dev-storage \
    -e TOKEN_SECRET=dev-secret \
    -v ~/.config/gcloud:/home/appuser/.config/gcloud:ro \
    cloud-docs-server:latest
```

### Using Docker Compose

```bash
# Copy environment template
cp .env.example .env

# Edit .env with your settings
# BUCKET_NAME=yourcompany-cloud-docs-dev-storage
# TOKEN_SECRET=your-secure-secret

# Run with Compose
docker-compose up -d

# View logs
docker-compose logs -f
```

## Static Asset Organization

The platform uses a two-tier approach for serving content:

### Document Structure
```
your-documents/
├── index.html              # Requires token
├── docs/                   # Requires token
│   ├── guide.html
│   └── reference.html
└── static/                 # No token required
    ├── main.css
    ├── app.js
    └── images/
        └── logo.png
```

### HTML Development
```html
<!DOCTYPE html>
<html>
<head>
    <!-- Static assets work with standard syntax -->
    <link rel="stylesheet" href="static/main.css">
    <script src="static/app.js"></script>
</head>
<body>
    <!-- Document content is token-protected -->
    <h1>Your Protected Content</h1>
</body>
</html>
```

**Benefits**:
- Standard web development practices work
- Better browser caching for CSS/JS assets  
- Easier integration with existing HTML documents
- Maintained security for sensitive document content

## Configuration

### Environment variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `BUCKET_NAME` | GCS bucket name | Required |
| `TOKEN_SECRET` | Secret for token signing | Required |
| `DOCS_PATH` | URL path prefix | `/docs` |
| `LOG_LEVEL` | Logging level | `info` |

### Cloud Run configuration

The application is configured to:
- Run as non-root user for security
- Use minimal Alpine Linux base image
- Include health checks for reliability
- Auto-scale based on traffic
- Use Cloud Run's built-in load balancing

## Security considerations

1. **Token Secret**: Use a cryptographically secure random string (32+ characters)
2. **IAM Permissions**: Service account only needs Storage Object Viewer permission
3. **HTTPS**: Cloud Run provides HTTPS by default
4. **Authentication**: All document endpoints require valid tokens
5. **Container Security**: Runs as non-root user with minimal attack surface

## Troubleshooting

### Common issues

1. **Build fails**: Ensure all dependencies are in go.mod and buildable
2. **Authentication errors**: Check GCP credentials and IAM permissions
3. **404 errors**: Verify bucket name and file paths
4. **Token errors**: Ensure TOKEN_SECRET matches between server and token generation

### Debugging

```bash
# Check service logs
gcloud run services logs read cloud-docs-server --region=us-central1

# Test individual components
./bin/token -validate YOUR_TOKEN
./bin/iframe -document /test.html -base-url YOUR_SERVICE_URL

# Check bucket contents
gsutil ls gs://yourcompany-cloud-docs-prod-storage/**
```

### Performance tuning

```bash
# Adjust concurrency and memory
gcloud run services update cloud-docs-server \
    --region=us-central1 \
    --concurrency=100 \
    --memory=512Mi \
    --cpu=1
```

## Cost optimization

- Cloud Storage: Pay per GB stored and requests
- Cloud Run: Pay per request and compute time
- Container Registry: Pay per GB stored
- Consider Cloud CDN for high-traffic deployments

For detailed pricing, see [GCP pricing documentation](https://cloud.google.com/pricing).