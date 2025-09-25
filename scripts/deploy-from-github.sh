#!/bin/bash

# Deploy directly from GitHub repository without local build
#
# WHEN TO USE:
#   - CI/CD pipelines
#   - Deploying from a clean state
#   - When you don't have the code locally
#   - Sharing deployment with team members
#
# PROS: No local setup needed, reproducible from any machine
# CONS: Requires public repo or GitHub authentication setup
#
# Example: ./deploy-from-github.sh cloud-docs-469520 my-bucket my-secret https://github.com/user/cloud-docs

set -e

# Configuration
PROJECT_ID=${1:-""}
BUCKET_NAME=${2:-""}
TOKEN_SECRET=${3:-""}
GITHUB_REPO=${4:-""}
REGION=${5:-"us-central1"}
SERVICE_NAME="cloud-docs-server"

# Check required parameters
if [ -z "$PROJECT_ID" ] || [ -z "$BUCKET_NAME" ] || [ -z "$TOKEN_SECRET" ] || [ -z "$GITHUB_REPO" ]; then
    echo "Usage: $0 <PROJECT_ID> <BUCKET_NAME> <TOKEN_SECRET> <GITHUB_REPO> [REGION]"
    echo ""
    echo "Example:"
    echo "  $0 my-project cloud-docs-bucket my-secret https://github.com/username/cloud-docs us-central1"
    echo ""
    echo "Note: Make sure your repository is public or you've set up authentication"
    exit 1
fi

echo "Deploying from GitHub to Cloud Run..."
echo "Project: $PROJECT_ID"
echo "GitHub Repo: $GITHUB_REPO"
echo "Region: $REGION"
echo ""

# Set the project
gcloud config set project $PROJECT_ID

# Deploy directly from source repository
echo "Deploying from GitHub repository..."
gcloud run deploy $SERVICE_NAME \
    --source $GITHUB_REPO \
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