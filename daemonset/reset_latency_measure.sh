#!/usr/bin/env bash
#
# Remove all “latency-measure” resources from the cluster.
# Optional: remove local docker image(s) whose repository matches IMAGE_NAME.
#
# Usage
#   ./reset_latency_measure.sh
#   CLEAN_IMAGE=true ./reset_latency_measure.sh   # also delete local image
#
set -euo pipefail

NAMESPACE="${NAMESPACE:-kube-system}"
IMAGE_NAME="${IMAGE_NAME:-latent-scheduler}"   # for optional docker rmi

echo "▶ Deleting DaemonSet / Deployment / Pods labelled app=latency-measure …"
kubectl delete Daemonset   latency-measure -n "$NAMESPACE" --ignore-not-found
kubectl delete deployment  latency-measure -n "$NAMESPACE" --ignore-not-found

# Anything else with the label in that namespace
kubectl delete all -n "$NAMESPACE" -l app=latency-measure --ignore-not-found

echo "▶ Deleting ServiceAccount + RBAC …"
kubectl delete serviceaccount              latency-measure-sa -n "$NAMESPACE"       --ignore-not-found
kubectl delete clusterrole                 latency-measure-role                     --ignore-not-found
kubectl delete clusterrolebinding          latency-measure-binding                  --ignore-not-found

echo "▶ Resources removed."

if [[ "${CLEAN_IMAGE:-false}" == "true" ]]; then
  echo "▶ Deleting local Docker images matching *${IMAGE_NAME}* (ignore errors if none) …"
  docker images | awk -v name="$IMAGE_NAME" 'NR>1 && $1 ~ name {print $1 ":" $2}' | xargs -r docker rmi -f
  echo "▶ Local images removed."
fi

echo "✅  Cluster reset complete."