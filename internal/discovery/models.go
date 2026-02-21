package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
)

// Provider represents an LLM provider from models.dev/api.json
type Provider struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Env    []string         `json:"env"`
	API    string           `json:"api"`
	Models map[string]Model `json:"models"`
}

// Model represents a specific LLM model.
type Model struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Registry handles model discovery and caching.
type Registry struct {
	cachePath string
	Providers map[string]Provider
}

// NewRegistry creates a new Registry with a default cache path.
func NewRegistry(cacheDir string) *Registry {
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".cache", "blunderbuss")
	}
	return &Registry{
		cachePath: filepath.Join(cacheDir, "models-api.json"),
		Providers: make(map[string]Provider),
	}
}

// GetCachePath returns the path to the models-api.json cache file.
func (r *Registry) GetCachePath() string {
	return r.cachePath
}

// Refresh fetches the latest api.json from models.dev and updates the cache.
func (r *Registry) Refresh() error {
	resp, err := http.Get("https://models.dev/api.json")
	if err != nil {
		return fmt.Errorf("fetching models.dev/api.json: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status from models.dev: %s", resp.Status)
	}

	var providers map[string]Provider
	if err := json.NewDecoder(resp.Body).Decode(&providers); err != nil {
		return fmt.Errorf("decoding api.json: %w", err)
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(r.cachePath), 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	data, err := json.MarshalIndent(providers, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling api.json for cache: %w", err)
	}

	if err := os.WriteFile(r.cachePath, data, 0644); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	r.Providers = providers
	return nil
}

// Load attempts to load providers from the local cache.
func (r *Registry) Load() error {
	data, err := os.ReadFile(r.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return r.Refresh() // Fetch if missing
		}
		return fmt.Errorf("reading cache file: %w", err)
	}

	if err := json.Unmarshal(data, &r.Providers); err != nil {
		return fmt.Errorf("unmarshaling cache file: %w", err)
	}

	return nil
}

// GetActiveModels returns a list of model IDs from providers that have their required env vars set.
func (r *Registry) GetActiveModels() []string {
	var activeModels []string
	for _, p := range r.Providers {
		if r.isProviderActive(p) {
			for _, m := range p.Models {
				// Format as provider/model-id
				activeModels = append(activeModels, fmt.Sprintf("%s/%s", p.ID, m.ID))
			}
		}
	}
	sort.Strings(activeModels)
	return activeModels
}

func (r *Registry) isProviderActive(p Provider) bool {
	if len(p.Env) == 0 {
		return false
	}
	for _, envVar := range p.Env {
		if os.Getenv(envVar) == "" {
			return false
		}
	}
	return true
}

// GetModelsForProvider returns a list of model IDs for a specific provider.
func (r *Registry) GetModelsForProvider(providerID string) []string {
	p, ok := r.Providers[providerID]
	if !ok {
		return nil
	}
	var models []string
	for _, m := range p.Models {
		models = append(models, fmt.Sprintf("%s/%s", p.ID, m.ID))
	}
	sort.Strings(models)
	return models
}
