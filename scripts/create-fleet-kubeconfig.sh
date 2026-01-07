#!/usr/bin/env bash
set -euo pipefail

AXON_KUBECONFIG=$(kind get kubeconfig --name axon)

FORMAT='{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}'
AXON_IP=$(docker inspect -f "$FORMAT" axon-control-plane)

AXON_KUBECONFIG=$(echo "$AXON_KUBECONFIG" | sed "s|https://127.0.0.1:[0-9]*|https://${AXON_IP}:6443|g")

kubectl --context kind-cortex create namespace dev 2>/dev/null || true

kubectl --context kind-cortex -n dev create secret generic cluster-kubeconfig \
    --from-literal=value="$AXON_KUBECONFIG" \
    --dry-run=client -o yaml | kubectl --context kind-cortex apply -f -

echo "Kubeconfig secret created in dev namespace"
