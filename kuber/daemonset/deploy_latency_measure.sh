#!/usr/bin/env bash
#
# Build → Push → Deploy the latency-measure DaemonSet end-to-end.
# It first wipes previous resources by calling reset_latency_measure.sh.
#
set -euo pipefail

### ─── User-tunable parameters (export to override) ──────────────────────────
REGISTRY="${REGISTRY:-sanjeevkrmenon}"      # docker hub user or private registry
IMAGE_NAME="${IMAGE_NAME:-latent-scheduler}"
NAMESPACE="${NAMESPACE:-kube-system}"
DOCKER_BIN="${DOCKER_BIN:-docker}"          # or podman / nerdctl
### ───────────────────────────────────────────────────────────────────────────

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "▶ Cleaning any previous install …"
"$SCRIPT_DIR/reset_latency_measure.sh"

TAG="daemon_$(date +%Y%m%d%H%M%S)"
FULL_IMAGE="${REGISTRY}/${IMAGE_NAME}:${TAG}"

echo "▶ Building image ${FULL_IMAGE}"
$DOCKER_BIN build -t "${FULL_IMAGE}" "${SCRIPT_DIR}"

echo "▶ Pushing image to registry"
$DOCKER_BIN push "${FULL_IMAGE}"

echo "▶ Applying RBAC objects"
kubectl apply -f "$SCRIPT_DIR/latency-rbac.yaml"

echo "▶ Rendering & applying DaemonSet"
sed "s|IMAGE_PLACEHOLDER|${FULL_IMAGE}|g" \
    "$SCRIPT_DIR/latency-measure-ds.yaml" | \
    kubectl apply -f -

echo "▶ Waiting for DaemonSet rollout"
kubectl rollout status daemonset/latency-measure -n "${NAMESPACE}"

echo "✅  Deployment finished.  You can watch logs with:"
echo "    kubectl logs -f -n ${NAMESPACE} -l app=latency-measure"