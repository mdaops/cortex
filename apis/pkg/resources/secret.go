package resources

import (
	kubeobj "github.com/crossplane-contrib/provider-kubernetes/apis/cluster/object/v1alpha2"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// SecretConfig configures a Secret.
type SecretConfig struct {
	Name            string
	Namespace       string
	Labels          map[string]string
	Annotations     map[string]string
	StringData      map[string]string
	Type            corev1.SecretType
	ProviderCfgName string
}

func (c *SecretConfig) withDefaults() {
	if c.ProviderCfgName == "" {
		c.ProviderCfgName = DefaultProviderCfg
	}
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
	if c.Type == "" {
		c.Type = corev1.SecretTypeOpaque
	}
}

// NewSecret creates a provider-kubernetes Object that manages a Secret.
func NewSecret(cfg SecretConfig) *kubeobj.Object {
	cfg.withDefaults()

	return &kubeobj.Object{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubernetes.crossplane.io/v1alpha2",
			Kind:       "Object",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.Namespace + "-" + cfg.Name + "-secret",
		},
		Spec: kubeobj.ObjectSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{Name: cfg.ProviderCfgName},
			},
			ForProvider: kubeobj.ObjectParameters{
				Manifest: runtime.RawExtension{
					Object: &corev1.Secret{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "v1",
							Kind:       "Secret",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:        cfg.Name,
							Namespace:   cfg.Namespace,
							Labels:      cfg.Labels,
							Annotations: cfg.Annotations,
						},
						StringData: cfg.StringData,
						Type:       cfg.Type,
					},
				},
			},
		},
	}
}
