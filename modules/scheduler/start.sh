#!/bin/bash

# Build and push Docker image
docker build -t sanjeevkrmenon/latent-scheduler:scheduler1 .
docker push sanjeevkrmenon/latent-scheduler:scheduler1

# Set file paths (we're already in the scheduler directory)
SCHEDULER_DEPLOYMENT=./scheduler-deployment.yaml
RBAC_CONFIG=./scheduler-rbac.yaml

# Apply RBAC and deploy the updated scheduler
kubectl apply -f $RBAC_CONFIG
kubectl apply -f $SCHEDULER_DEPLOYMENT
