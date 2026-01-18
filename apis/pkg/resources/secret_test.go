package resources

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestNewSecret(t *testing.T) {
	tests := []struct {
		name     string
		cfg      SecretConfig
		wantName string
	}{
		{
			name: "basic",
			cfg: SecretConfig{
				Name:      "artifacts-s3",
				Namespace: "finance",
				StringData: map[string]string{
					"accessKey": "test-key",
					"secretKey": "test-secret",
				},
			},
			wantName: "finance-artifacts-s3-secret",
		},
		{
			name: "custom type",
			cfg: SecretConfig{
				Name:      "artifacts-s3",
				Namespace: "finance",
				Type:      corev1.SecretTypeOpaque,
				StringData: map[string]string{
					"accessKey": "test-key",
				},
			},
			wantName: "finance-artifacts-s3-secret",
		},
		{
			name: "custom provider",
			cfg: SecretConfig{
				Name:            "artifacts-s3",
				Namespace:       "finance",
				ProviderCfgName: "custom",
				StringData: map[string]string{
					"accessKey": "test-key",
				},
			},
			wantName: "finance-artifacts-s3-secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret := NewSecret(tt.cfg)
			if secret.Name != tt.wantName {
				t.Errorf("expected name %s, got %s", tt.wantName, secret.Name)
			}
			if secret.Spec.ForProvider.Manifest.Object == nil {
				t.Fatal("expected manifest object")
			}
			if tt.cfg.ProviderCfgName != "" && secret.Spec.ProviderConfigReference.Name != tt.cfg.ProviderCfgName {
				t.Errorf("expected provider %s, got %s", tt.cfg.ProviderCfgName, secret.Spec.ProviderConfigReference.Name)
			}
		})
	}
}
