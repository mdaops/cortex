package resources

import (
	"errors"
	"fmt"

	argocd "github.com/crossplane-contrib/provider-argocd/apis/projects/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultArgoCDNamespace is the default namespace where ArgoCD is installed.
const DefaultArgoCDNamespace = "argo-system"

// DefaultClusterServer is the default Kubernetes API server URL for in-cluster.
const DefaultClusterServer = "https://kubernetes.default.svc"

// ErrSourceReposRequired is returned when sourceRepos is empty.
var ErrSourceReposRequired = errors.New("sourceRepos is required and cannot be empty")

// ProjectConfig configures an ArgoCD project for a tenant.
type ProjectConfig struct {
	TenantName             string
	Description            string
	SourceRepos            []string
	ArgoCDNamespace        string
	ProviderCfgName        string
	AdditionalDestinations []argocd.ApplicationDestination
}

func (c *ProjectConfig) withDefaults() {
	if c.ArgoCDNamespace == "" {
		c.ArgoCDNamespace = DefaultArgoCDNamespace
	}
	if c.ProviderCfgName == "" {
		c.ProviderCfgName = DefaultProviderCfg
	}
	if c.Description == "" {
		c.Description = fmt.Sprintf("%s tenant workloads", c.TenantName)
	}
}

func (c *ProjectConfig) validate() error {
	if len(c.SourceRepos) == 0 {
		return ErrSourceReposRequired
	}
	return nil
}

// NewProject creates a provider-argocd Project with the given configuration.
func NewProject(cfg ProjectConfig) (*argocd.Project, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	cfg.withDefaults()

	server := DefaultClusterServer
	wildcardNs := cfg.TenantName + "-*"

	destinations := []argocd.ApplicationDestination{
		{
			Namespace: &cfg.TenantName,
			Server:    &server,
		},
		{
			Namespace: &wildcardNs,
			Server:    &server,
		},
	}
	destinations = append(destinations, cfg.AdditionalDestinations...)

	return &argocd.Project{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "projects.argocd.crossplane.io/v1alpha1",
			Kind:       "Project",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.TenantName + "-project",
		},
		Spec: argocd.ProjectSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{Name: cfg.ProviderCfgName},
			},
			ForProvider: argocd.ProjectParameters{
				Description:  &cfg.Description,
				SourceRepos:  cfg.SourceRepos,
				Destinations: destinations,
			},
		},
	}, nil
}

// NewTenantProject is a convenience function for creating a tenant ArgoCD project.
func NewTenantProject(tenantName, description string, sourceRepos []string) (*argocd.Project, error) {
	return NewProject(ProjectConfig{
		TenantName:  tenantName,
		Description: description,
		SourceRepos: sourceRepos,
	})
}
