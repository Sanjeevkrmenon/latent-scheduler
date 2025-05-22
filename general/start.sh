#!/bin/bash
docker build -t sanjeevkrmenon/latent-scheduler:w1 .
docker push sanjeevkrmenon/latent-scheduler:w1

SCHEDULER_DEPLOYMENT=scheduler-deployment.yaml
RBAC_CONFIG=scheduler-rbac.yaml


SCHEDULER_DEPLOYMENT=./kuber/scheduler-deployment.yaml
RBAC_CONFIG=./kuber/scheduler-rbac.yaml

# Apply RBAC and deploy the updated scheduler
kubectl apply -f $RBAC_CONFIG
kubectl apply -f $SCHEDULER_DEPLOYMENT
