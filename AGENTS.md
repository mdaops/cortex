# Agent Instructions for Cortex

Control plane for Synapse, a Kubernetes-native MLOps platform.
Cortex (hub) manages Axon (spoke) clusters via Flux GitOps.

## Structure

```
cortex/                     # Hub cluster - YOU ARE HERE
├── hub/                    # Flux Kustomizations (self + axon management)
├── deploy/                 # Manifests deployed to cortex
│   ├── tenants/            # Namespaces, RBAC for managed clusters
│   ├── infra-controllers/  # CRD controllers (Crossplane, Kyverno)
│   └── infra-configs/      # Cluster-wide custom resources
├── clusters/               # Kind cluster configurations
├── scripts/                # Shell scripts
├── flake.nix               # Nix dev environment
└── Justfile                # Task runner

axon/                       # Spoke cluster (sibling repo)
├── deploy/{tenants,infra-controllers,infra-configs,apps}/
└── flake.nix
```

## Commands

```bash
nix develop                   # Enter dev shell with all tools

just fleet-up                 # Create Kind clusters
just fleet-down               # Destroy clusters
just fleet-status             # Check health

just bootstrap mdaops cortex  # Bootstrap Flux (needs GITHUB_TOKEN)
just flux-status              # View Flux resources
just flux-reconcile           # Force sync
just flux-watch               # Watch kustomizations

just validate                 # Validate manifests with kubeconform
just ctx-cortex               # Switch to cortex context
just ctx-axon                 # Switch to axon context
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

Flux dependency chain: tenants -> infra-controllers -> infra-configs -> apps

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

## Git

Commit messages: imperative mood, no period
- Good: `Add Crossplane provider configuration`
- Bad: `Added crossplane provider configuration.`

## Testing

1. Validate before commit: `just validate`
2. After push, check Flux: `just flux-status`
3. Verify target: `kubectl --context kind-axon get pods -A`

## Architecture

- Cortex runs Flux controllers
- Axon kubeconfig stored as Secret in `axon` namespace on cortex
- Flux applies to axon via `kubeConfig.secretRef`
- Changes to axon/ repo pulled by cortex, applied remotely

## Do Not

- Add comments to YAML, Justfile, or scripts
- Use `latest` image tags
- Commit secrets or kubeconfig files
- Skip validation before pushing
- Use emojis in log output
