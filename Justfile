set dotenv-load
set shell := ["bash", "-uc"]

default:
    @just --list

fleet-up:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Creating Cortex cluster..."
    kind create cluster --config clusters/cortex.yaml --name cortex 2>/dev/null || echo "Cortex cluster already exists"
    echo "Creating Axon cluster..."
    kind create cluster --config clusters/axon.yaml --name axon 2>/dev/null || echo "Axon cluster already exists"
    echo "Creating kubeconfig secret for Axon..."
    just _create-axon-kubeconfig
    echo ""
    echo "Fleet is up"
    just fleet-status

fleet-down:
    #!/usr/bin/env bash
    echo "Destroying fleet..."
    kind delete cluster --name cortex 2>/dev/null || true
    kind delete cluster --name axon 2>/dev/null || true
    echo "Fleet destroyed"

fleet-restart: fleet-down fleet-up

fleet-status:
    #!/usr/bin/env bash
    echo "Fleet Status"
    echo ""
    echo "=== Cortex (Hub) ==="
    if kubectl --context kind-cortex get nodes &>/dev/null; then
        kubectl --context kind-cortex get nodes
    else
        echo "  Not running"
    fi
    echo ""
    echo "=== Axon (Spoke) ==="
    if kubectl --context kind-axon get nodes &>/dev/null; then
        kubectl --context kind-axon get nodes
    else
        echo "  Not running"
    fi

bootstrap owner repo:
    #!/usr/bin/env bash
    set -euo pipefail
    if [ -z "${GITHUB_TOKEN:-}" ]; then
        echo "Error: GITHUB_TOKEN is required"
        echo "  export GITHUB_TOKEN=<your-token>"
        exit 1
    fi
    echo "Bootstrapping Flux on Cortex..."
    flux bootstrap github \
        --context=kind-cortex \
        --owner={{owner}} \
        --repository={{repo}} \
        --branch=main \
        --path=hub \
        --personal
    echo "Flux bootstrapped"

flux-reconcile:
    flux reconcile source git flux-system --context kind-cortex
    flux reconcile kustomization flux-system --context kind-cortex --with-source

flux-watch:
    watch -n2 flux get kustomizations -A --context kind-cortex

flux-status:
    @echo "=== Sources ==="
    @flux get sources git -A --context kind-cortex
    @echo ""
    @echo "=== Kustomizations ==="
    @flux get kustomizations -A --context kind-cortex
    @echo ""
    @echo "=== Helm Releases ==="
    @flux get helmreleases -A --context kind-cortex 2>/dev/null || echo "No HelmReleases"

flux-suspend:
    flux suspend kustomization --all --context kind-cortex

flux-resume:
    flux resume kustomization --all --context kind-cortex

validate:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Validating Cortex manifests..."
    for dir in deploy/*/; do
        if [ -f "$dir/kustomization.yaml" ]; then
            echo "  Checking $dir..."
            kustomize build "$dir" | kubeconform -strict -skip CustomResourceDefinition
        fi
    done
    echo "Validating Axon manifests..."
    for dir in ../axon/deploy/*/; do
        if [ -f "$dir/kustomization.yaml" ]; then
            echo "  Checking $dir..."
            kustomize build "$dir" | kubeconform -strict -skip CustomResourceDefinition
        fi
    done
    echo ""
    echo "All manifests valid"

_create-axon-kubeconfig:
    ./scripts/create-axon-kubeconfig.sh

k9s-cortex:
    k9s --context kind-cortex

k9s-axon:
    k9s --context kind-axon

ctx-cortex:
    kubectl config use-context kind-cortex

ctx-axon:
    kubectl config use-context kind-axon

shell:
    nix develop

update:
    nix flake update
