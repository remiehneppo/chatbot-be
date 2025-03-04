#!/bin/bash
set -e

# Build the image
docker compose build

# Save the image to a tar file
docker save weaviate-weaviate > weaviate-offline.tar

# Create a zip file containing everything needed
zip -r weaviate-portable.zip \
    docker-compose.yml \
    Dockerfile \
    weaviate-offline.tar \
    README.md