# Cloud Docs Scripts

## Deployment Scripts

| Script | Use Case | Speed | Requirements |
|--------|----------|-------|--------------|
| `deploy.sh` | Production deployments | ⚡ Slow | gcloud CLI only |
| `deploy-podman.sh` | Development/testing | ⚡⚡⚡ Fast | Podman + gcloud |
| `deploy-from-github.sh` | CI/CD, team deploys | ⚡⚡ Medium | gcloud CLI only |

### Quick Start

```bash
# Production deployment (safest, slowest)
./deploy.sh PROJECT_ID BUCKET TOKEN_SECRET

# Fast local deployment with Podman (development)
./deploy-podman.sh PROJECT_ID BUCKET TOKEN_SECRET

# Deploy from GitHub (CI/CD friendly)
./deploy-from-github.sh PROJECT_ID BUCKET TOKEN_SECRET GITHUB_URL
```

## Document Processing Scripts

| Script | Purpose |
|--------|---------|
| `convert_all.sh` | Convert Markdown to HTML with embedded tokens |
| `add_token.lua` | Pandoc filter to inject tokens into links/images |
| `copy_btn.html` | HTML snippet for copy button functionality |
| `puppeteer-config.json` | Mermaid CLI configuration |
| `set-token.sh` | Helper to set token in environment |

### Document Conversion Workflow

```bash
# Generate token
TOKEN=$(../cmd/token/token -g -e 87600h)

# Convert all documentation
./convert_all.sh -s ../courses -d ../output -t "$TOKEN" -c ../static

# Upload to Cloud Storage
gcloud storage rsync -r ../output/courses/ gs://BUCKET/courses/
gcloud storage rsync -r ../output/static/ gs://BUCKET/static/
```

## Script Selection Guide

**Choose `deploy.sh` when:**
- Deploying to production
- You don't have Docker/Podman locally
- You want the most reliable, tested path

**Choose `deploy-podman.sh` when:**
- Doing rapid development/testing
- You have Podman installed
- You want fastest iteration time
- You need to test the container locally first

**Choose `deploy-from-github.sh` when:**
- Setting up CI/CD
- Deploying from a different machine
- Ensuring reproducible deployments
- Working with a team

## Prerequisites

### For all scripts
- gcloud CLI installed and configured
- Appropriate GCP permissions

### For deploy-podman.sh only
```bash
# Install Podman on macOS
brew install podman

# Initialize and start Podman machine
podman machine init
podman machine start
```

### For convert_all.sh
```bash
# Install conversion tools
brew install pandoc
brew install coreutils  # For gmktemp on macOS
npm install -g @mermaid-js/mermaid-cli
```