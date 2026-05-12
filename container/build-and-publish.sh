#!/bin/bash
set -e

IMAGE="starwhenry/rinha-backend-2026"
TAG=$(date +%Y%m%d%H%M)
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
ROOT_DIR=$(cd "$SCRIPT_DIR/.." && pwd)

docker build \
    -t "$IMAGE:$TAG" \
    -t "$IMAGE:latest" \
    -t "rinha-backend-2026:latest" \
    -f "$SCRIPT_DIR/Dockerfile" "$ROOT_DIR"

docker push --all-tags $IMAGE

echo "Published $IMAGE:$TAG"