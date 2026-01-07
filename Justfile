# Cortex - Synapse Platform Control Plane

set dotenv-load
set shell := ["bash", "-uc"]

default:
    @just --list

# ==================== Fleet Lifecycle ====================

# Create both Kind clusters
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

# Destroy both clusters
fleet-down:
    #!/usr/bin/env bash
    echo "Destroying fleet..."
    kind delete cluster --name cortex 2>/dev/null || true
    kind delete cluster --name axon 2>/dev/null || true
    echo "Fleet destroyed"

# Restart fleet
fleet-restart: fleet-down fleet-up

# Check fleet health
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

# ==================== Bootstrap ====================

# Bootstrap Flux on Cortex
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

# ==================== Flux Operations ====================

# Force reconciliation
flux-reconcile:
    flux reconcile source git flux-system --context kind-cortex
    flux reconcile kustomization flux-system --context kind-cortex --with-source

# Watch all kustomizations
flux-watch:
    watch -n2 flux get kustomizations -A --context kind-cortex

# Get Flux status
flux-status:
    @echo "=== Sources ==="
    @flux get sources git -A --context kind-cortex
    @echo ""
    @echo "=== Kustomizations ==="
    @flux get kustomizations -A --context kind-cortex
    @echo ""
    @echo "=== Helm Releases ==="
    @flux get helmreleases -A --context kind-cortex 2>/dev/null || echo "No HelmReleases"

# Suspend Flux reconciliation
flux-suspend:
    flux suspend kustomization --all --context kind-cortex

# Resume Flux reconciliation
flux-resume:
    flux resume kustomization --all --context kind-cortex

# ==================== Validation ====================

# Validate all manifests
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

# ==================== Utilities ====================

# Create kubeconfig secret for Axon management
_create-axon-kubeconfig:
    #!/usr/bin/env bash
    set -euo pipefail

    # Get the axon kubeconfig
    AXON_KUBECONFIG=$(kind get kubeconfig --name axon)

    # Get the axon container IP (Kind uses container networking)
    AXON_IP=$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' axon-control-plane)

    # Replace localhost with container IP
    AXON_KUBECONFIG=$(echo "$AXON_KUBECONFIG" | sed "s|https://127.0.0.1:[0-9]*|https://${AXON_IP}:6443|g")

    # Create namespace if not exists
    kubectl --context kind-cortex create namespace axon 2>/dev/null || true

    # Create/update secret
    kubectl --context kind-cortex -n axon create secret generic cluster-kubeconfig \
        --from-literal=value="$AXON_KUBECONFIG" \
        --dry-run=client -o yaml | kubectl --context kind-cortex apply -f -

    echo "Kubeconfig secret created in axon namespace"

# Interactive k9s for cortex
k9s-cortex:
    k9s --context kind-cortex

# Interactive k9s for axon
k9s-axon:
    k9s --context kind-axon

# Switch kubectl context to cortex
ctx-cortex:
    kubectl config use-context kind-cortex

# Switch kubectl context to axon
ctx-axon:
    kubectl config use-context kind-axon

# Enter nix dev shell
shell:
    nix develop

# Update flake inputs
update:
    nix flake update
