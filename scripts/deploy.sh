#!/bin/bash

# Primary deployment script using Google Cloud Build
#
# WHEN TO USE:
#   - Production deployments (most reliable)
#   - When you want Google to handle the build
#   - When local Docker/Podman is not available
#
# PROS: Simple, no local container runtime needed
# CONS: Slower (uploads source, waits for Cloud Build)
#
# Example: ./deploy.sh cloud-docs-469520 my-bucket my-secret us-central1

set -e

# Configuration
PROJECT_ID=${1:-""}
BUCKET_NAME=${2:-""}
TOKEN_SECRET=${3:-""}
REGION=${4:-"us-central1"}
SERVICE_NAME="cloud-docs-server"

# Check required parameters
if [ -z "$PROJECT_ID" ] || [ -z "$BUCKET_NAME" ] || [ -z "$TOKEN_SECRET" ]; then
    echo "Usage: $0 <PROJECT_ID> <BUCKET_NAME> <TOKEN_SECRET> [REGION]"
    echo ""
    echo "Example:"
    echo "  $0 my-project cloud-docs-storage-bucket my-secure-secret us-central1"
    echo ""
    echo "Required:"
    echo "  PROJECT_ID    - Your GCP project ID"
    echo "  BUCKET_NAME   - Your Cloud Storage bucket name"
    echo "  TOKEN_SECRET  - Secure secret for token signing"
    echo ""
    echo "Optional:"
    echo "  REGION        - GCP region (default: us-central1)"
    exit 1
fi

echo "Deploying to Google Cloud Run..."
echo "Project: $PROJECT_ID"
echo "Bucket: $BUCKET_NAME"
echo "Region: $REGION"
echo "Service: $SERVICE_NAME"
echo ""

# Set the project
gcloud config set project $PROJECT_ID

# Build and deploy using Cloud Build
gcloud builds submit . \
    --config cloudbuild.yaml \
    --substitutions _BUCKET_NAME=$BUCKET_NAME,_TOKEN_SECRET=$TOKEN_SECRET,_REGION=$REGION

echo ""
echo "Deployment complete!"
echo ""
echo "To get the service URL:"
echo "gcloud run services describe $SERVICE_NAME --region=$REGION --format='value(status.url)'"

# Get and display the service URL
SERVICE_URL=$(gcloud run services describe $SERVICE_NAME --region=$REGION --format='value(status.url)' 2>/dev/null || echo "")
if [ -n "$SERVICE_URL" ]; then
    echo ""
    echo "Service URL: $SERVICE_URL"
    echo "Health check: $SERVICE_URL/health"
fi