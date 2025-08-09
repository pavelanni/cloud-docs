#!/bin/bash

# Local build script using Podman
set -e

PROJECT_ID=${1:-"cloud-docs-project"}
IMAGE_NAME="cloud-docs-server"
TAG=${2:-"latest"}

echo "Building container image locally with Podman..."

# Build the image
podman build -t ${IMAGE_NAME}:${TAG} .

echo "Build complete!"
echo "Image: ${IMAGE_NAME}:${TAG}"

# Test the image locally
echo ""
echo "To test locally:"
echo "podman run -p 8080:8080 -e BUCKET_NAME=your-bucket -e TOKEN_SECRET=your-secret ${IMAGE_NAME}:${TAG}"

# Tag for GCR if PROJECT_ID is provided
if [ "$PROJECT_ID" != "cloud-docs-project" ]; then
    GCR_TAG="gcr.io/${PROJECT_ID}/${IMAGE_NAME}:${TAG}"
    echo ""
    echo "Tagging for Google Container Registry..."
    podman tag ${IMAGE_NAME}:${TAG} ${GCR_TAG}
    echo "Tagged as: ${GCR_TAG}"
    echo ""
    echo "To push to GCR:"
    echo "podman push ${GCR_TAG}"
fi