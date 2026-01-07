# Cortex

Synapse Platform Control Plane - the hub cluster for MLOps.

## Overview

Cortex is the central management cluster that orchestrates the Synapse platform. It runs Flux and manages the Axon product cluster(s) via GitOps.

## Quick Start

```bash
# Enter dev shell
nix develop

# Create clusters
just fleet-up

# Bootstrap Flux (requires GITHUB_TOKEN)
export GITHUB_TOKEN=<your-token>
just bootstrap mdaops cortex

# Watch reconciliation
just flux-watch
```

## Structure

```
cortex/
├── clusters/          # Kind cluster definitions
├── deploy/            # Base Kustomize definitions
│   ├── tenants/       # Namespace and RBAC for managed clusters
│   ├── infra-controllers/  # CRD controllers (Crossplane, Kyverno, etc.)
│   └── infra-configs/      # Cluster-wide custom resources
└── hub/               # Flux configuration
    ├── cortex.yaml    # Self-management
    └── axon.yaml      # Spoke management
```

## Commands

| Command | Description |
|---------|-------------|
| `just fleet-up` | Create Kind clusters |
| `just fleet-down` | Destroy clusters |
| `just fleet-status` | Check cluster health |
| `just bootstrap <owner> <repo>` | Bootstrap Flux |
| `just flux-status` | View Flux status |
| `just flux-reconcile` | Force reconciliation |
| `just validate` | Validate manifests |
