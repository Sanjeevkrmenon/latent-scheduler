#!/usr/bin/env bash
#
# Build -> Push -> Deploy the latency-measure DaemonSet using a fixed image tag.
#

set -euo pipefail

# ---- User-tunable parameters (or override with env) ----
IMAGE_NAME="${IMAGE_NAME:-sanjeevkrmenon/latency-measure:latest}"  # <--- hardcoded!
NAMESPACE="${NAMESPACE:-kube-system}"
DOCKER_BIN="${DOCKER_BIN:-docker}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
# --------------------------------------------------------

if [[ ! -x "$SCRIPT_DIR/reset_latency_measure.sh" ]]; then
    echo "Error: reset_latency_measure.sh not found or not executable!" >&2
    exit 1
fi 

echo "▶ Cleaning any previous install …"
"$SCRIPT_DIR/reset_latency_measure.sh"

echo "▶ Building image ${IMAGE_NAME} ..."
$DOCKER_BIN build -t "${IMAGE_NAME}" "${SCRIPT_DIR}"

echo "▶ Pushing image to registry ..."
$DOCKER_BIN push "${IMAGE_NAME}"

echo "▶ Applying RBAC objects ..."
kubectl apply -f "$SCRIPT_DIR/latency-rbac.yaml"

echo "▶ Applying DaemonSet manifest ..."
kubectl apply -f "$SCRIPT_DIR/latency-measure-ds.yaml"

echo "▶ Waiting for DaemonSet rollout ..."
kubectl rollout status daemonset/latency-measure -n "${NAMESPACE}"

echo
echo "✅ Deployment finished!"
echo "   You can stream logs from all nodes with:"
echo "       kubectl logs -f -n ${NAMESPACE} -l app=latency-measure"