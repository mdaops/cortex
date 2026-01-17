package resources

import (
	"testing"
)

func TestNewTenantNamespace(t *testing.T) {
	ns := NewTenantNamespace("finance")

	if ns.Name != "finance-namespace" {
		t.Errorf("expected name finance-namespace, got %s", ns.Name)
	}
	if ns.Spec.ForProvider.Manifest.Object == nil {
		t.Fatal("expected manifest object")
	}
}

func TestNewNamespace(t *testing.T) {
	tests := []struct {
		name   string
		cfg    NamespaceConfig
		wantNs string
	}{
		{
			name:   "basic",
			cfg:    NamespaceConfig{Name: "test"},
			wantNs: "test-namespace",
		},
		{
			name: "custom labels",
			cfg: NamespaceConfig{
				Name:   "test",
				Labels: map[string]string{"env": "prod"},
			},
			wantNs: "test-namespace",
		},
		{
			name: "custom provider",
			cfg: NamespaceConfig{
				Name:            "test",
				ProviderCfgName: "custom",
			},
			wantNs: "test-namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := NewNamespace(tt.cfg)
			if ns.Name != tt.wantNs {
				t.Errorf("expected name %s, got %s", tt.wantNs, ns.Name)
			}
			if tt.cfg.ProviderCfgName != "" && ns.Spec.ProviderConfigReference.Name != tt.cfg.ProviderCfgName {
				t.Errorf("expected provider %s, got %s", tt.cfg.ProviderCfgName, ns.Spec.ProviderConfigReference.Name)
			}
		})
	}
}
