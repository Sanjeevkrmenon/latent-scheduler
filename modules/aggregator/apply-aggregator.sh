#!/bin/bash

set -e

NAMESPACE="kube-system"
AGGREGATOR_MANIFEST="latency-aggregator-cronjob.yaml"
DOCKERFILE_PATH="."    # Path to directory with Dockerfile and aggregator.py
IMAGE_NAME="sanjeevkrmenon/latency-aggregator:latest"

# Optional: declare other manifests
# PVC_MANIFEST="latency-pvc.yaml"
# DAEMONSET_MANIFEST="latency-measure-daemonset.yaml"

echo "==== Building aggregator Docker image ===="
docker build -t $IMAGE_NAME $DOCKERFILE_PATH

echo "==== Pushing aggregator Docker image to registry ===="
docker push $IMAGE_NAME

echo "==== Applying aggregator CronJob ===="
kubectl apply -n $NAMESPACE -f "$AGGREGATOR_MANIFEST"

# Optional: If you want to re-apply PVC or DaemonSet, uncomment these:
# echo "Applying PersistentVolumeClaim..."
# kubectl apply -n $NAMESPACE -f "$PVC_MANIFEST"

# echo "Applying latency measurement DaemonSet..."
# kubectl apply -n $NAMESPACE -f "$DAEMONSET_MANIFEST"

echo "==== Resources applied! Current status: ===="
kubectl -n $NAMESPACE get cronjobs,jobs,pods | grep aggregator

echo "==== Done. ===="