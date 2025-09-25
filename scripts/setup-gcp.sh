#!/bin/bash

# GCP project setup script
set -e

PROJECT_ID=${1:-""}
BUCKET_SUFFIX=${2:-""}

if [ -z "$PROJECT_ID" ]; then
    echo "Usage: $0 <PROJECT_ID> [BUCKET_SUFFIX]"
    echo ""
    echo "Example:"
    echo "  $0 my-cloud-docs-project 2025"
    echo ""
    echo "This will create:"
    echo "  - Project: my-cloud-docs-project"
    echo "  - Bucket: cloud-docs-storage-2025"
    echo ""
    exit 1
fi

BUCKET_NAME="cloud-docs-storage-${BUCKET_SUFFIX:-$PROJECT_ID}"
REGION="us-central1"

echo "Setting up GCP project for Cloud Docs..."
echo "Project ID: $PROJECT_ID"
echo "Bucket Name: $BUCKET_NAME"
echo "Region: $REGION"
echo ""

# Check if project exists and create if it doesn't
if ! gcloud projects describe $PROJECT_ID &>/dev/null; then
    echo "Creating project: $PROJECT_ID"
    gcloud projects create $PROJECT_ID
    echo "Project created successfully!"
else
    echo "Project $PROJECT_ID already exists."
fi

# Set the project
gcloud config set project $PROJECT_ID

#echo ""
#echo "Enabling required APIs..."

# Enable required APIs
#gcloud services enable storage.googleapis.com
#gcloud services enable run.googleapis.com
#gcloud services enable cloudbuild.googleapis.com
#gcloud services enable containerregistry.googleapis.com

#echo "APIs enabled successfully!"
#echo ""

# Create Cloud Storage bucket
echo "Creating Cloud Storage bucket: $BUCKET_NAME"
if ! gcloud storage buckets list $BUCKET_NAME > /dev/null; then
    gcloud storage buckets create $BUCKET_NAME --location=$REGION
    echo "Bucket created successfully!"
else
    echo "Bucket $BUCKET_NAME already exists."
fi

# Set up bucket permissions (optional - for public access)
echo ""
echo "Bucket setup complete!"
echo ""
echo "Next steps:"
echo "1. Upload your documents to gs://$BUCKET_NAME using:"
echo "   ./bin/upload -source ./docs -bucket $BUCKET_NAME"
echo ""
echo "2. Generate a secure token secret:"
echo "   TOKEN_SECRET=\$(openssl rand -base64 32)"
echo ""
echo "3. Deploy the application:"
echo "   ./scripts/deploy.sh $PROJECT_ID $BUCKET_NAME \$TOKEN_SECRET"
echo ""
echo "4. Test the deployment:"
echo "   # Get service URL"
echo "   SERVICE_URL=\$(gcloud run services describe cloud-docs-server --region=$REGION --format='value(status.url)')"
echo "   # Test health check"
echo "   curl \$SERVICE_URL/health"