package resources

import (
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	kubeobj "github.com/crossplane-contrib/provider-kubernetes/apis/cluster/object/v1alpha2"
)

// TenantLabelKey is the label key used to identify tenant resources.
const TenantLabelKey = "platform.synapse.io/tenant"

// DefaultProviderCfg is the default provider config name.
const DefaultProviderCfg = "default"

// NamespaceConfig configures a tenant namespace.
type NamespaceConfig struct {
	Name            string
	Labels          map[string]string
	Annotations     map[string]string
	ProviderCfgName string
}

func (c *NamespaceConfig) withDefaults() {
	if c.ProviderCfgName == "" {
		c.ProviderCfgName = DefaultProviderCfg
	}
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	c.Labels[TenantLabelKey] = c.Name
}

// NewNamespace creates a provider-kubernetes Object that manages a Namespace.
func NewNamespace(cfg NamespaceConfig) *kubeobj.Object {
	cfg.withDefaults()

	return &kubeobj.Object{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubernetes.crossplane.io/v1alpha2",
			Kind:       "Object",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.Name + "-namespace",
		},
		Spec: kubeobj.ObjectSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{Name: cfg.ProviderCfgName},
			},
			ForProvider: kubeobj.ObjectParameters{
				Manifest: runtime.RawExtension{
					Object: &corev1.Namespace{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "v1",
							Kind:       "Namespace",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:        cfg.Name,
							Labels:      cfg.Labels,
							Annotations: cfg.Annotations,
						},
					},
				},
			},
		},
	}
}

// NewTenantNamespace is a convenience function for creating a simple tenant namespace.
func NewTenantNamespace(tenantName string) *kubeobj.Object {
	return NewNamespace(NamespaceConfig{Name: tenantName})
}
