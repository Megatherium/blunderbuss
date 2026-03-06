package discovery

import (
	"fmt"
	"os"
	"sort"
)

// GetActiveModels returns a list of model IDs from providers that have their required env vars set.
func (r *Registry) GetActiveModels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var activeModels []string
	for _, provider := range r.providers {
		if !r.isProviderActive(provider) {
			continue
		}
		activeModels = append(activeModels, formatProviderModels(provider)...)
	}

	sort.Strings(activeModels)
	return activeModels
}

func (r *Registry) isProviderActive(provider Provider) bool {
	for _, envVar := range provider.Env {
		if os.Getenv(envVar) == "" {
			return false
		}
	}
	return true
}

// GetModelsForProvider returns a list of model IDs for a specific provider.
func (r *Registry) GetModelsForProvider(providerID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, ok := r.providers[providerID]
	if !ok {
		return nil
	}

	models := formatProviderModels(provider)
	sort.Strings(models)
	return models
}

func formatProviderModels(provider Provider) []string {
	models := make([]string, 0, len(provider.Models))
	for _, model := range provider.Models {
		models = append(models, fmt.Sprintf("%s/%s", provider.ID, model.ID))
	}
	return models
}

// SetProviders is a helper method used for tests to inject mock providers.
func (r *Registry) SetProviders(providers map[string]Provider) {
	r.setProviders(providers)
}

func (r *Registry) setProviders(providers map[string]Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers = providers
}
