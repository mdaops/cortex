package resources

import (
	"errors"
	"testing"
)

func TestNewProject(t *testing.T) {
	tests := []struct {
		name    string
		cfg     ProjectConfig
		wantErr error
	}{
		{
			name: "valid config",
			cfg: ProjectConfig{
				TenantName:  "finance",
				SourceRepos: []string{"https://github.com/org/repo.git"},
			},
			wantErr: nil,
		},
		{
			name: "missing sourceRepos",
			cfg: ProjectConfig{
				TenantName: "finance",
			},
			wantErr: ErrSourceReposRequired,
		},
		{
			name: "empty sourceRepos",
			cfg: ProjectConfig{
				TenantName:  "finance",
				SourceRepos: []string{},
			},
			wantErr: ErrSourceReposRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proj, err := NewProject(tt.cfg)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && proj == nil {
				t.Error("expected project, got nil")
			}
		})
	}
}

func TestNewTenantProject(t *testing.T) {
	proj, err := NewTenantProject("finance", "Finance team", []string{"https://github.com/org/repo.git"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if proj.Name != "finance-project" {
		t.Errorf("expected name finance-project, got %s", proj.Name)
	}
	if *proj.Spec.ForProvider.Description != "Finance team" {
		t.Errorf("expected description Finance team, got %s", *proj.Spec.ForProvider.Description)
	}
	if len(proj.Spec.ForProvider.Destinations) != 2 {
		t.Errorf("expected 2 destinations, got %d", len(proj.Spec.ForProvider.Destinations))
	}
}

func TestNewTenantProject_RequiresSourceRepos(t *testing.T) {
	_, err := NewTenantProject("finance", "desc", nil)
	if !errors.Is(err, ErrSourceReposRequired) {
		t.Errorf("expected ErrSourceReposRequired, got %v", err)
	}

	_, err = NewTenantProject("finance", "desc", []string{})
	if !errors.Is(err, ErrSourceReposRequired) {
		t.Errorf("expected ErrSourceReposRequired, got %v", err)
	}
}

func TestProjectConfig_DefaultDescription(t *testing.T) {
	proj, err := NewProject(ProjectConfig{
		TenantName:  "myteam",
		SourceRepos: []string{"*"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *proj.Spec.ForProvider.Description != "myteam tenant workloads" {
		t.Errorf("expected default description, got %s", *proj.Spec.ForProvider.Description)
	}
}
