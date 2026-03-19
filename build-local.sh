#!/bin/bash
set -e

VERSION=$(cat VERSION 2>/dev/null || echo "0.1.0")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "Building GSBE v${VERSION} (${COMMIT})"

# Install npm dependencies if needed
if [ ! -d "node_modules" ]; then
    echo "Installing npm dependencies..."
    npm install
fi

# Build Tailwind CSS
echo "Building Tailwind CSS..."
npm run build:css

# Build Docker image
echo "Building Docker image..."
docker build \
    --no-cache \
    --build-arg VERSION="${VERSION}" \
    --build-arg BUILD_DATE="${BUILD_DATE}" \
    --build-arg COMMIT="${COMMIT}" \
    -t gsbe:${VERSION} \
    -t gsbe:latest \
    -t gsbe:local \
    .

echo "Build complete:"
echo "  - gsbe:${VERSION}"
echo "  - gsbe:latest"
echo "  - gsbe:local"
