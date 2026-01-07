#!/usr/bin/env bash
set -euo pipefail

kubectl --context kind-cortex create namespace dev 2>/dev/null || true

kind get kubeconfig --internal --name axon > /tmp/axon.kubeconfig

kubectl --context kind-cortex -n dev create secret generic cluster-kubeconfig \
    --from-file=value=/tmp/axon.kubeconfig \
    --dry-run=client -o yaml | kubectl --context kind-cortex apply -f -

rm /tmp/axon.kubeconfig

echo "Kubeconfig secret created in dev namespace"
