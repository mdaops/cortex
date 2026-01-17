# Agent Instructions for Cortex

Control plane for Synapse, a Kubernetes-native MLOps platform.
Cortex manages platform infrastructure and fleet clusters via Flux GitOps.
Argo CD (deployed by Flux) handles application workloads on fleet clusters.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                       CORTEX (Platform Cluster)                     │
│                                                                     │
│  Flux manages:                                                      │
│  ├── Platform infrastructure (cert-manager, kyverno, etc.)         │
│  ├── Argo CD installation & configuration on fleet clusters        │
│  ├── Tenant boundaries (namespaces, quotas, RBAC)                  │
│  └── Control plane components across all clusters                  │
└─────────────────────────────────────────────────────────────────────┘
                              │
                              │ Deploys control plane TO
                              ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      FLEET CLUSTERS (Spokes)                        │
│                                                                     │
│  Argo CD manages:                                                   │
│  ├── Application workloads (Deployments, Services, etc.)           │
│  ├── Product team resources (Buckets, ML pipelines, etc.)          │
│  └── Application-level configs                                      │
└─────────────────────────────────────────────────────────────────────┘
```

## Structure

```
cortex/
├── hub/                    # Flux entry point (bootstrapped to cortex)
│   ├── flux-system/        # Flux components (auto-generated)
│   ├── cortex.yaml         # Platform cluster Kustomizations
│   └── dev.yaml            # Dev fleet cluster Kustomizations
│
├── clusters/               # Per-cluster overlays
│   ├── cortex/             # Platform cluster
│   │   ├── controllers/    # Platform-specific controller config
│   │   ├── config/
│   │   └── tenants/
│   ├── dev/                # Fleet: dev environment
│   │   ├── controllers/    # → components/controllers + patches
│   │   ├── config/
│   │   ├── gateway/        # Gateway CRDs + controller
│   │   └── tenants/
│   └── production/         # Fleet: production environment
│       ├── controllers/
│       ├── config/
│       └── tenants/
│
├── components/             # Shared base definitions
│   ├── controllers/        # HelmReleases: argo-cd, cert-manager, etc.
│   ├── config/             # Policies, gateway routes, tenant templates
│   │   ├── policies/
│   │   ├── gateway/
│   │   └── tenant/
│   ├── crds/               # CRD-only installs
│   │   └── gateway-api/
│   ├── crossplane/         # Crossplane deployment artifacts
│   │   ├── configurations/
│   │   └── providerconfigs/
│   └── tenants/
│
├── apis/                   # Crossplane composition packages
│   └── management/         # Platform API definitions
│       ├── package/
│       └── functions/
│
├── kind/                   # Kind cluster configurations (local dev)
├── scripts/                # Shell scripts
├── flake.nix               # Nix dev environment
└── Justfile                # Task runner
```

## Flux Dependency Chain

1. `tenants` - Namespaces, service accounts, role bindings
2. `controllers` - CRD controllers (depends on tenants)
3. `config` - Cluster-wide custom resources (depends on controllers)

## Commands

```bash
nix develop                   # Enter dev shell with all tools

just fleet-up                 # Create Kind clusters
just fleet-down               # Destroy clusters
just fleet-status             # Check health

just bootstrap mdaops cortex  # Bootstrap Flux (needs GITHUB_TOKEN)
just flux-status              # View Flux resources
just flux-reconcile           # Force sync

just validate                 # Validate manifests
```

## YAML Style

Multi-document files use `---` separator:
```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: example
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: example
  namespace: example
```

Field order: apiVersion, kind, metadata, spec

Always specify:
- `namespace` for namespaced resources
- `resources.requests` and `resources.limits` for containers
- Explicit image tags (never `latest`)

Kustomization files - keep minimal:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - resource-a.yaml
```

## Nix Style

Group packages by purpose in buildInputs:
```nix
buildInputs = with pkgs; [
  kubectl
  kustomize
  just
];
```

## Shell Scripts

Strict mode always:
```bash
#!/usr/bin/env bash
set -euo pipefail
```

No comments. Self-documenting code. Echo status for feedback.

## Justfile Style

No comments. Self-documenting recipe names.
Shebang for multi-line bash:
```just
recipe-name:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Doing something..."
```

## File Naming

- Manifests: `lowercase-with-dashes.yaml`
- Scripts: `lowercase-with-dashes.sh`
- Kustomization: always `kustomization.yaml`

## Commits

Conventional commits only. Format: `type: description`

Types:
- `feat`: new feature or capability
- `fix`: bug fix
- `docs`: documentation only
- `refactor`: code change that neither fixes nor adds
- `chore`: maintenance, dependencies, tooling
- `ci`: CI/CD changes
- `test`: adding or updating tests

Rules:
- Lowercase type and description
- No period at end
- Imperative mood
- Max 72 characters

Examples:
- `feat: add crossplane provider configuration`
- `fix: correct kubeconfig secret path`
- `chore: update flux to v2.2.0`
- `refactor: simplify fleet-up script`

## Testing

1. Validate before commit: `just validate`
2. After push, check Flux: `just flux-status`
3. Verify fleet cluster: `kubectl --context kind-axon get pods -A`

## Do Not

- Add comments to YAML, Justfile, or scripts
- Use `latest` image tags
- Commit secrets or kubeconfig files
- Skip validation before pushing
- Use emojis in log output
