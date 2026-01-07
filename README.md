# Cortex

Synapse Platform Control Plane - manages fleet clusters via Flux GitOps.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                       CORTEX (Platform Cluster)                     │
│                                                                     │
│  Flux manages:                                                      │
│  ├── Platform infrastructure (Crossplane, Kyverno, etc.)           │
│  ├── Argo CD installation on fleet clusters                        │
│  └── Tenant boundaries (namespaces, quotas, RBAC)                  │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ Deploys control plane TO
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      FLEET CLUSTERS (Spokes)                        │
│                                                                     │
│  Argo CD manages:                                                   │
│  ├── Application workloads                                          │
│  └── Product team resources                                         │
└─────────────────────────────────────────────────────────────────────┘
```

## Quick Start

```bash
nix develop

just fleet-up

export GITHUB_TOKEN=<your-token>
just bootstrap mdaops cortex

just flux-watch
```

## Structure

```
cortex/
├── hub/                    # Flux entry point
│   ├── flux-system/        # Flux controllers
│   ├── cortex.yaml         # Platform Kustomizations
│   └── dev.yaml            # Fleet dev Kustomizations
│
├── deploy/                 # Base definitions (shared)
│   ├── controllers/        # HelmReleases
│   ├── config/             # Policies, configs
│   └── tenants/            # SAs, RBAC, namespaces
│
├── platform/               # Platform cluster overlays
│   ├── controllers/
│   ├── config/
│   └── tenants/
│
├── fleet/                  # Fleet cluster overlays
│   ├── dev/
│   │   ├── controllers/
│   │   ├── config/
│   │   └── tenants/
│   └── production/
│       └── ...
│
└── kind/                   # Kind cluster configs
    ├── cortex.yaml
    └── axon.yaml
```

## Dependency Chain

Flux applies in order: `tenants` → `controllers` → `config`

Each namespace has its own set:
- `flux-system/tenants` → `flux-system/controllers` → `flux-system/config` (platform)
- `dev/tenants` → `dev/controllers` → `dev/config` (fleet dev)

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

## Adding a Fleet Environment

1. Create overlay in `fleet/<env>/`
2. Create hub file `hub/<env>.yaml`
3. Add to `hub/kustomization.yaml`
4. Create kubeconfig secret in `<env>` namespace on cortex
