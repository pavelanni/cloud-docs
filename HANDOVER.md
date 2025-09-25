# Cloud Docs Authentication Handover

## Overview

This system uses token-based authentication to protect documentation from public access while allowing authorized access through LMS embeds.

## Secrets Location

All secrets are stored in **1Password** in the "Cloud Docs Infrastructure" vault:

- **TOKEN_SECRET**: Master secret for token generation/validation (DO NOT CHANGE unless rotating all tokens)
- **TOKEN**: Pre-generated 10-year token for HTML embedding and LMS

## Where Secrets Are Used

### 1. Cloud Run Server (GCP Console)

- Project: `cloud-docs-469520`
- Service: `cloud-docs-server`
- Environment Variable: `TOKEN_SECRET`
- Region: `us-central1`

### 2. GitHub Actions (Documentation Repository)

- Repository: [your-org/documentation]
- Location: Settings > Secrets and variables > Actions
- Secret name: `DOCS_TOKEN` (the generated token, not TOKEN_SECRET)

### 3. LMS Iframe Embeds

- All iframe src URLs contain `?token=` parameter
- Same token used across all embeds
- Manual update process - maintain list of course IDs

## Token Management

### Current Token Information

- Token Generated: [DATE]
- Token Expires: [DATE + 10 years]
- Calendar Reminder Set: [DATE + 9 years] (1 year warning)

### Generate New Token

```bash
# Build the token tool
cd cmd/token
go build -o token

# Generate a 10-year token
TOKEN=$(./token -g -e 87600h)
echo "$TOKEN"

# Validate token details
./token -v "$TOKEN"
```

### Test Token Validity

```bash
# Quick test with curl
curl "https://cloud-docs-server.run.app/health"  # Should work without token
curl "https://cloud-docs-server.run.app/docs/test.html?token=[TOKEN]"  # Should return content
curl "https://cloud-docs-server.run.app/docs/test.html"  # Should return 401
```

## Emergency Token Rotation

If TOKEN_SECRET is compromised or needs rotation:

1. **Generate new TOKEN_SECRET**
   ```bash
   # Generate a secure random secret
   openssl rand -base64 32
   ```

2. **Update Cloud Run environment variable**
   - Go to GCP Console > Cloud Run > cloud-docs-server
   - Edit & Deploy New Revision
   - Update TOKEN_SECRET environment variable
   - Deploy

3. **Generate new TOKEN using new secret**
   ```bash
   export TOKEN_SECRET="new-secret-here"
   TOKEN=$(./token -g -e 87600h)
   ```

4. **Re-convert all documentation**
   ```bash
   ./scripts/convert_all.sh -s courses/ -d output/ -t "$TOKEN"
   ```

5. **Re-upload to Cloud Storage**
   ```bash
   gcloud storage rsync -r --delete-unmatched-destination-objects \
     ./output/courses/ gs://cloud-docs-469520/courses/
   ```

6. **Update GitHub Actions secret**
   - Update DOCS_TOKEN in repository secrets

7. **Update all LMS embeds** (most time-consuming)
   - This is manual - check LMS course list

## Build and Deployment

### Local Development

```bash
# Build all tools
make build

# Run server locally
export TOKEN_SECRET="your-secret"
export BUCKET_NAME="cloud-docs-469520"
./dist/darwin-arm64/server
```

### Deployment Options

We maintain three deployment scripts for different scenarios:

#### 1. Production Deployment (Cloud Build)
**Script:** `scripts/deploy.sh`
**When to use:** Production releases, when you want Google to handle everything
**Pros:** Simple, reliable, no local container runtime needed
**Cons:** Slower (uploads source, waits for Cloud Build queue)

```bash
./scripts/deploy.sh \
  cloud-docs-469520 \
  cloud-docs-469520 \
  "$TOKEN_SECRET" \
  us-central1
```

#### 2. Fast Development Deployment (Podman + Artifact Registry)
**Script:** `scripts/deploy-podman.sh`
**When to use:** Development/testing, rapid iteration
**Pros:** Fastest builds, uses modern Artifact Registry, can test locally
**Cons:** Requires Podman installed locally

```bash
# First time: ensure Podman is installed
brew install podman
podman machine init
podman machine start

# Deploy
./scripts/deploy-podman.sh \
  cloud-docs-469520 \
  cloud-docs-469520 \
  "$TOKEN_SECRET" \
  us-central1
```

#### 3. GitHub Repository Deployment
**Script:** `scripts/deploy-from-github.sh`
**When to use:** CI/CD, deploying from clean state, team deployments
**Pros:** No local setup needed, reproducible from any machine
**Cons:** Requires public repo or GitHub auth setup

```bash
./scripts/deploy-from-github.sh \
  cloud-docs-469520 \
  cloud-docs-469520 \
  "$TOKEN_SECRET" \
  https://github.com/pavelanni/cloud-docs
```

## Documentation Conversion Workflow

### Prerequisites

```bash
# Install required tools
npm install -g @mermaid-js/mermaid-cli
brew install pandoc  # macOS
brew install coreutils  # macOS (for gmktemp)
```

### Convert and Upload

```bash
# Convert Markdown to HTML with embedded tokens
./scripts/convert_all.sh \
  -s courses/ \
  -d output/ \
  -t "$TOKEN" \
  -c /path/to/static/assets

# Upload to Cloud Storage
gcloud storage rsync -r --delete-unmatched-destination-objects \
  ./output/courses/ gs://cloud-docs-469520/courses/
gcloud storage rsync -r --delete-unmatched-destination-objects \
  ./output/static/ gs://cloud-docs-469520/static/
```

## Important Notes

1. **Token Philosophy**: We use long-lived (10-year) tokens because:
   - LMS embedding is manual and rotation is expensive
   - Goal is preventing indexing, not high security
   - Content is educational, not highly sensitive

2. **Security Model**:
   - Tokens prevent search engine indexing
   - Tokens prevent casual public access
   - NOT designed to prevent determined attackers
   - Students could share URLs (accepted risk)

3. **Backup Critical Items**:
   - TOKEN_SECRET in 1Password
   - Generated TOKEN in 1Password
   - This documentation in repo
   - List of LMS courses using embeds

4. **Monitoring**:
   - Check Cloud Run logs for authentication failures
   - Set up uptime monitoring for health endpoint
   - Calendar reminder 1 year before token expiration

## Contacts

- Cloud Docs Owner: [Your Name]
- GCP Project Owner: [Project Owner]
- LMS Administrator: [LMS Admin]
- 1Password Vault Access: [Team/Group Name]

## Related Documentation

- [README.md](README.md) - General project overview
- [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) - Detailed deployment instructions
- [CLAUDE.md](CLAUDE.md) - AI assistant instructions for codebase