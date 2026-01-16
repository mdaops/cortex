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

## Tailscale Integration

Fleet cluster services are exposed via Tailscale for secure access without port-forwarding.

### Prerequisites

1. Create OAuth client at https://login.tailscale.com/admin/settings/oauth
   - Scopes: `devices:write`, `dns:write`
   - Tag: `tag:k8s` (or your preferred tag)

2. Create the secret on cortex cluster before bootstrapping:

```bash
kubectl --context kind-cortex create namespace dev
kubectl --context kind-cortex create secret generic tailscale-oauth \
  --namespace dev \
  --from-literal=clientId=<your-client-id> \
  --from-literal=clientSecret=<your-client-secret>
```

### Custom Domain Access

Services are exposed via kgateway with TLS termination on custom domains.

**Architecture:**
```
Client → DNS → Tailscale (encrypted) → LoadBalancer → kgateway (TLS) → Service
```

Or use an A record pointing to the Tailscale IP (check with `kubectl get svc synapse-gateway -n kgateway-system`).

**Certificate:**

Development uses a self-signed CA. Browser will show a certificate warning.

For production, configure Let's Encrypt with DNS-01 challenge:

1. Create a ClusterIssuer with your DNS provider credentials
2. Update `deploy/config/gateway/issuer.yaml` to use Let's Encrypt
3. Update `deploy/config/gateway/certificate.yaml` issuer reference

## Adding a Fleet Environment

1. Create overlay in `fleet/<env>/`
2. Create hub file `hub/<env>.yaml`
3. Add to `hub/kustomization.yaml`
4. Create kubeconfig secret in `<env>` namespace on cortex
