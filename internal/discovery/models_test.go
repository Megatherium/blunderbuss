// Copyright (C) 2026 megatherium
package discovery

import (
	"os"
	"testing"
)

func TestIsProviderActive(t *testing.T) {
	registry := NewRegistry("")
	provider := Provider{
		ID:  "test-provider",
		Env: []string{"TEST_API_KEY"},
	}

	// Should be inactive initially
	os.Unsetenv("TEST_API_KEY")
	if registry.isProviderActive(provider) {
		t.Errorf("expected provider to be inactive when env var is missing")
	}

	// Should be active when env var is set
	os.Setenv("TEST_API_KEY", "dummy")
	defer os.Unsetenv("TEST_API_KEY")
	if !registry.isProviderActive(provider) {
		t.Errorf("expected provider to be active when env var is set")
	}
}

func TestGetActiveModels(t *testing.T) {
	registry := NewRegistry("")
	registry.Providers = map[string]Provider{
		"p1": {
			ID: "p1",
			Env: []string{"P1_KEY"},
			Models: map[string]Model{
				"m1": {ID: "m1", Name: "Model 1"},
			},
		},
		"p2": {
			ID: "p2",
			Env: []string{"P2_KEY"},
			Models: map[string]Model{
				"m2": {ID: "m2", Name: "Model 2"},
			},
		},
	}

	os.Setenv("P1_KEY", "val")
	defer os.Unsetenv("P1_KEY")
	os.Unsetenv("P2_KEY")

	active := registry.GetActiveModels()
	if len(active) != 1 {
		t.Fatalf("expected 1 active model, got %d", len(active))
	}

	if active[0] != "p1/m1" {
		t.Errorf("expected active model p1/m1, got %s", active[0])
	}
}
