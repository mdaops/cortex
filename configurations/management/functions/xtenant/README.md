# function-xtenant

Crossplane composition function for tenant provisioning. Creates a Kubernetes Namespace and ArgoCD AppProject for each Tenant CR.

## Architecture

```
configurations/
├── pkg/
│   └── resources/              # Shared typed resource builders
│       ├── namespace.go        # NewTenantNamespace()
│       ├── project.go          # NewTenantProject()
│       └── convert.go          # ConvertViaJSON()
│
└── management/
    └── functions/
        └── xtenant/
            ├── fn.go           # Orchestration only
            ├── fn_test.go      # Tests
            └── main.go         # gRPC entrypoint
```

## Usage

Create a Tenant claim:

```yaml
apiVersion: platform.synapse.io/v1alpha1
kind: Tenant
metadata:
  name: finance
  namespace: default
spec:
  name: finance
  description: Finance team workloads
  sourceRepos:
    - https://github.com/org/finance-apps.git
```

The function creates:
- Kubernetes Namespace `finance` with label `platform.synapse.io/tenant: finance`
- ArgoCD AppProject `finance` scoped to namespace `finance` and `finance-*`

## Pattern: Shared Typed Resource Builders

Resource builders live in a shared module (`configurations/pkg/resources/`) so multiple composition functions can reuse them.

### Function Code

```go
import "github.com/mdaops/cortex/configurations/pkg/resources"

desiredTyped := map[resource.Name]any{
    "namespace": resources.NewTenantNamespace(tenantName),
    "project":   resources.NewTenantProject(tenantName, description, sourceRepos),
}

for name, obj := range desiredTyped {
    c := composed.New()
    if err := resources.ConvertViaJSON(c, obj); err != nil {
        return fatal(err)
    }
    desired[name] = &resource.DesiredComposed{Resource: c}
}
```

### Adding New Resources

1. Add builder to shared module (`configurations/pkg/resources/`):

```go
func NewTenantServiceAccount(tenantName string) *kubeobj.Object {
    return &kubeobj.Object{
        TypeMeta: metav1.TypeMeta{
            APIVersion: "kubernetes.crossplane.io/v1alpha2",
            Kind:       "Object",
        },
        // ... fully typed
    }
}
```

2. Use in any function:

```go
desiredTyped["service-account"] = resources.NewTenantServiceAccount(tenantName)
```

### Benefits

| Aspect | Untyped (old) | Typed + Shared (new) |
|--------|---------------|----------------------|
| Type safety | None | Full compile-time |
| IDE support | None | Autocomplete |
| Code reuse | Copy-paste | Import |
| Adding resources | ~30 lines | ~5 lines |
| Cross-function sharing | Manual | Automatic |

## Development

```bash
go test ./...

docker build . --platform=linux/amd64 --tag=runtime-amd64

crossplane xpkg build --package-root=package --embed-runtime-image=runtime-amd64 --package-file=function.xpkg
```

## References

- [Crossplane Composition Functions](https://docs.crossplane.io/latest/concepts/composition-functions)
- [Upbound Go Composition Guide](https://docs.upbound.io/manuals/cli/howtos/compositions/go)
- [function-sdk-go](https://pkg.go.dev/github.com/crossplane/function-sdk-go)
