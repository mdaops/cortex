# Crossplane Resource Builders

Shared typed resource builders for Crossplane composition functions.

## Usage

```go
import "github.com/mdaops/cortex/configurations/pkg/resources"

// In go.mod, add replace directive for local development:
// replace github.com/mdaops/cortex/configurations/pkg/resources => ../../../pkg/resources

namespace := resources.NewTenantNamespace("finance")
project, err := resources.NewTenantProject("finance", "Finance team", []string{"https://github.com/org/repo.git"})
if err != nil {
    return err  // sourceRepos is required
}

// Convert to unstructured for Crossplane SDK
c := composed.New()
if err := resources.ConvertViaJSON(c, namespace); err != nil {
    return err
}
```

## Available Builders

### Namespaces

```go
// Simple tenant namespace with default labels
ns := resources.NewTenantNamespace("tenant-name")

// Customized namespace
ns := resources.NewNamespace(resources.NamespaceConfig{
    Name:            "tenant-name",
    Labels:          map[string]string{"env": "prod"},
    Annotations:     map[string]string{"owner": "team-a"},
    ProviderCfgName: "custom-provider",
})
```

### ArgoCD Projects

```go
// Simple tenant project (sourceRepos is required)
proj, err := resources.NewTenantProject("tenant-name", "description", []string{"https://github.com/org/repo"})

// Customized project
proj, err := resources.NewProject(resources.ProjectConfig{
    TenantName:      "tenant-name",
    Description:     "Custom description",
    SourceRepos:     []string{"https://github.com/org/repo"},
    ArgoCDNamespace: "argocd",  // default: argo-system
    ProviderCfgName: "custom",   // default: default
})
```

## Adding New Builders

1. Create a new file (e.g., `serviceaccount.go`)
2. Import the provider types
3. Create builder functions with config structs for flexibility

Example:

```go
package resources

import (
    kubeobj "github.com/crossplane-contrib/provider-kubernetes/apis/cluster/object/v1alpha2"
)

type ServiceAccountConfig struct {
    Name            string
    Namespace       string
    ProviderCfgName string
}

func NewServiceAccount(cfg ServiceAccountConfig) *kubeobj.Object {
    // ... implementation
}
```

## Design Principles

1. **Type Safety**: Use provider Go types, not `map[string]interface{}`
2. **Config Structs**: Use config structs for builders with many options
3. **Sensible Defaults**: Apply defaults via `withDefaults()` methods
4. **Convenience Functions**: Provide simple wrappers for common cases
