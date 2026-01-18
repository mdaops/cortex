package resources

import (
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"
)

func TestNewServiceAccount(t *testing.T) {
	tests := []struct {
		name     string
		cfg      ServiceAccountConfig
		wantName string
	}{
		{
			name: "basic",
			cfg: ServiceAccountConfig{
				Name:      "argo-workflow",
				Namespace: "finance",
			},
			wantName: "finance-argo-workflow-sa",
		},
		{
			name: "custom provider",
			cfg: ServiceAccountConfig{
				Name:            "argo-workflow",
				Namespace:       "finance",
				ProviderCfgName: "custom",
			},
			wantName: "finance-argo-workflow-sa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sa := NewServiceAccount(tt.cfg)
			if sa.Name != tt.wantName {
				t.Errorf("expected name %s, got %s", tt.wantName, sa.Name)
			}
			if sa.Spec.ForProvider.Manifest.Object == nil {
				t.Fatal("expected manifest object")
			}
			if tt.cfg.ProviderCfgName != "" && sa.Spec.ProviderConfigReference.Name != tt.cfg.ProviderCfgName {
				t.Errorf("expected provider %s, got %s", tt.cfg.ProviderCfgName, sa.Spec.ProviderConfigReference.Name)
			}
		})
	}
}

func TestNewRole(t *testing.T) {
	rules := []rbacv1.PolicyRule{
		{
			APIGroups: []string{"argoproj.io"},
			Resources: []string{"workflowtaskresults"},
			Verbs:     []string{"create", "patch"},
		},
	}

	tests := []struct {
		name     string
		cfg      RoleConfig
		wantName string
	}{
		{
			name: "basic",
			cfg: RoleConfig{
				Name:      "argo-workflow",
				Namespace: "finance",
				Rules:     rules,
			},
			wantName: "finance-argo-workflow-role",
		},
		{
			name: "custom provider",
			cfg: RoleConfig{
				Name:            "argo-workflow",
				Namespace:       "finance",
				Rules:           rules,
				ProviderCfgName: "custom",
			},
			wantName: "finance-argo-workflow-role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := NewRole(tt.cfg)
			if role.Name != tt.wantName {
				t.Errorf("expected name %s, got %s", tt.wantName, role.Name)
			}
			if role.Spec.ForProvider.Manifest.Object == nil {
				t.Fatal("expected manifest object")
			}
			if tt.cfg.ProviderCfgName != "" && role.Spec.ProviderConfigReference.Name != tt.cfg.ProviderCfgName {
				t.Errorf("expected provider %s, got %s", tt.cfg.ProviderCfgName, role.Spec.ProviderConfigReference.Name)
			}
		})
	}
}

func TestNewRoleBinding(t *testing.T) {
	tests := []struct {
		name     string
		cfg      RoleBindingConfig
		wantName string
	}{
		{
			name: "basic",
			cfg: RoleBindingConfig{
				Name:               "argo-workflow",
				Namespace:          "finance",
				RoleName:           "argo-workflow",
				ServiceAccountName: "argo-workflow",
			},
			wantName: "finance-argo-workflow-rolebinding",
		},
		{
			name: "custom provider",
			cfg: RoleBindingConfig{
				Name:               "argo-workflow",
				Namespace:          "finance",
				RoleName:           "argo-workflow",
				ServiceAccountName: "argo-workflow",
				ProviderCfgName:    "custom",
			},
			wantName: "finance-argo-workflow-rolebinding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rb := NewRoleBinding(tt.cfg)
			if rb.Name != tt.wantName {
				t.Errorf("expected name %s, got %s", tt.wantName, rb.Name)
			}
			if rb.Spec.ForProvider.Manifest.Object == nil {
				t.Fatal("expected manifest object")
			}
			if tt.cfg.ProviderCfgName != "" && rb.Spec.ProviderConfigReference.Name != tt.cfg.ProviderCfgName {
				t.Errorf("expected provider %s, got %s", tt.cfg.ProviderCfgName, rb.Spec.ProviderConfigReference.Name)
			}
		})
	}
}
