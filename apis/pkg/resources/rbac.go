package resources

import (
	kubeobj "github.com/crossplane-contrib/provider-kubernetes/apis/cluster/object/v1alpha2"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ServiceAccountConfig configures a ServiceAccount.
type ServiceAccountConfig struct {
	Name            string
	Namespace       string
	Labels          map[string]string
	Annotations     map[string]string
	ProviderCfgName string
}

func (c *ServiceAccountConfig) withDefaults() {
	if c.ProviderCfgName == "" {
		c.ProviderCfgName = DefaultProviderCfg
	}
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
}

// NewServiceAccount creates a provider-kubernetes Object that manages a ServiceAccount.
func NewServiceAccount(cfg ServiceAccountConfig) *kubeobj.Object {
	cfg.withDefaults()

	return &kubeobj.Object{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubernetes.crossplane.io/v1alpha2",
			Kind:       "Object",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.Namespace + "-" + cfg.Name + "-sa",
		},
		Spec: kubeobj.ObjectSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{Name: cfg.ProviderCfgName},
			},
			ForProvider: kubeobj.ObjectParameters{
				Manifest: runtime.RawExtension{
					Object: &corev1.ServiceAccount{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "v1",
							Kind:       "ServiceAccount",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:        cfg.Name,
							Namespace:   cfg.Namespace,
							Labels:      cfg.Labels,
							Annotations: cfg.Annotations,
						},
					},
				},
			},
		},
	}
}

// RoleConfig configures a Role.
type RoleConfig struct {
	Name            string
	Namespace       string
	Labels          map[string]string
	Rules           []rbacv1.PolicyRule
	ProviderCfgName string
}

func (c *RoleConfig) withDefaults() {
	if c.ProviderCfgName == "" {
		c.ProviderCfgName = DefaultProviderCfg
	}
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
}

// NewRole creates a provider-kubernetes Object that manages a Role.
func NewRole(cfg RoleConfig) *kubeobj.Object {
	cfg.withDefaults()

	return &kubeobj.Object{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubernetes.crossplane.io/v1alpha2",
			Kind:       "Object",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.Namespace + "-" + cfg.Name + "-role",
		},
		Spec: kubeobj.ObjectSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{Name: cfg.ProviderCfgName},
			},
			ForProvider: kubeobj.ObjectParameters{
				Manifest: runtime.RawExtension{
					Object: &rbacv1.Role{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "rbac.authorization.k8s.io/v1",
							Kind:       "Role",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      cfg.Name,
							Namespace: cfg.Namespace,
							Labels:    cfg.Labels,
						},
						Rules: cfg.Rules,
					},
				},
			},
		},
	}
}

// RoleBindingConfig configures a RoleBinding.
type RoleBindingConfig struct {
	Name               string
	Namespace          string
	Labels             map[string]string
	RoleName           string
	ServiceAccountName string
	ProviderCfgName    string
}

func (c *RoleBindingConfig) withDefaults() {
	if c.ProviderCfgName == "" {
		c.ProviderCfgName = DefaultProviderCfg
	}
	if c.Labels == nil {
		c.Labels = make(map[string]string)
	}
}

// NewRoleBinding creates a provider-kubernetes Object that manages a RoleBinding.
func NewRoleBinding(cfg RoleBindingConfig) *kubeobj.Object {
	cfg.withDefaults()

	return &kubeobj.Object{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "kubernetes.crossplane.io/v1alpha2",
			Kind:       "Object",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.Namespace + "-" + cfg.Name + "-rolebinding",
		},
		Spec: kubeobj.ObjectSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{Name: cfg.ProviderCfgName},
			},
			ForProvider: kubeobj.ObjectParameters{
				Manifest: runtime.RawExtension{
					Object: &rbacv1.RoleBinding{
						TypeMeta: metav1.TypeMeta{
							APIVersion: "rbac.authorization.k8s.io/v1",
							Kind:       "RoleBinding",
						},
						ObjectMeta: metav1.ObjectMeta{
							Name:      cfg.Name,
							Namespace: cfg.Namespace,
							Labels:    cfg.Labels,
						},
						RoleRef: rbacv1.RoleRef{
							APIGroup: "rbac.authorization.k8s.io",
							Kind:     "Role",
							Name:     cfg.RoleName,
						},
						Subjects: []rbacv1.Subject{
							{
								Kind:      "ServiceAccount",
								Name:      cfg.ServiceAccountName,
								Namespace: cfg.Namespace,
							},
						},
					},
				},
			},
		},
	}
}
