#!/bin/bash

set -euo pipefail

# ---- Config ----
IMAGE_NAME="sanjeevkrmenon/latent-scheduler:scheduler1"
SCHEDULER_DEPLOYMENT="./scheduler-deployment.yaml"
RBAC_CONFIG="./scheduler-rbac.yaml"
DOCKERFILE="."

# ---- Build and Push ----
echo "=== Building Docker image: $IMAGE_NAME ==="
docker build -t "$IMAGE_NAME" $DOCKERFILE

echo "=== Pushing Docker image: $IMAGE_NAME ==="
docker push "$IMAGE_NAME"

# ---- Kubernetes Apply ----
echo "=== Applying scheduler RBAC ==="
kubectl apply -f "$RBAC_CONFIG"

echo "=== Applying scheduler deployment ==="
kubectl apply -f "$SCHEDULER_DEPLOYMENT"

echo "=== Scheduler deployment/update complete! ==="