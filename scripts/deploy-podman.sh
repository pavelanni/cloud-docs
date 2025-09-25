#!/bin/bash

# Fast local deployment using Podman and Google Artifact Registry
#
# WHEN TO USE:
#   - Development and testing (fastest iteration)
#   - When you have Podman installed locally
#   - When you want to test the container before deploying
#
# PROS: Fast builds, modern Artifact Registry, local testing possible
# CONS: Requires Podman installed and configured
#
# Prerequisites:
#   - Podman installed (brew install podman)
#   - gcloud CLI configured
#
# Example: ./deploy-podman.sh cloud-docs-469520 my-bucket my-secret us-central1

set -e

# Configuration
PROJECT_ID=${1:-""}
BUCKET_NAME=${2:-""}
TOKEN_SECRET=${3:-""}
REGION=${4:-"us-central1"}
SERVICE_NAME="cloud-docs-server"
REGISTRY_NAME="cloud-docs"
IMAGE_NAME="$REGION-docker.pkg.dev/$PROJECT_ID/$REGISTRY_NAME/$SERVICE_NAME"

# Check required parameters
if [ -z "$PROJECT_ID" ] || [ -z "$BUCKET_NAME" ] || [ -z "$TOKEN_SECRET" ]; then
    echo "Usage: $0 <PROJECT_ID> <BUCKET_NAME> <TOKEN_SECRET> [REGION]"
    echo ""
    echo "Example:"
    echo "  $0 my-project cloud-docs-storage-bucket my-secure-secret us-central1"
    exit 1
fi

echo "Deploying via Artifact Registry (using Podman)..."
echo "Project: $PROJECT_ID"
echo "Registry: $REGISTRY_NAME"
echo "Image: $IMAGE_NAME"
echo ""

# Set the project
gcloud config set project $PROJECT_ID

# Create Artifact Registry repository if it doesn't exist
echo "Ensuring Artifact Registry repository exists..."
gcloud artifacts repositories create $REGISTRY_NAME \
    --repository-format=docker \
    --location=$REGION \
    --description="Cloud Docs container images" 2>/dev/null || echo "Repository already exists"

# Get authentication token for Artifact Registry
echo "Getting authentication token..."
ACCESS_TOKEN=$(gcloud auth print-access-token)

# Build the Go binary for Linux
echo "Building Go binary for Linux..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
    -ldflags="-s -w" \
    -o server ./cmd/server/

# Build container image with Podman
echo "Building container image with Podman..."
podman build -t $IMAGE_NAME .

# Login to Artifact Registry with Podman
echo "Authenticating Podman with Artifact Registry..."
echo $ACCESS_TOKEN | podman login -u oauth2accesstoken --password-stdin https://$REGION-docker.pkg.dev

# Push to Artifact Registry
echo "Pushing to Artifact Registry..."
podman push $IMAGE_NAME

# Deploy to Cloud Run
echo "Deploying to Cloud Run..."
gcloud run deploy $SERVICE_NAME \
    --image $IMAGE_NAME \
    --region $REGION \
    --platform managed \
    --allow-unauthenticated \
    --set-env-vars "BUCKET_NAME=$BUCKET_NAME,TOKEN_SECRET=$TOKEN_SECRET,DOCS_PATH=/docs" \
    --memory 512Mi \
    --cpu 1 \
    --timeout 300 \
    --max-instances 10

echo ""
echo "Deployment complete!"

# Get service URL
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region=$REGION --format='value(status.url)')
echo "Service URL: $SERVICE_URL"
echo "Health check: $SERVICE_URL/health"

# Clean up
rm -f server