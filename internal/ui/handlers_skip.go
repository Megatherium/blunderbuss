package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/megatherium/blunderbust/internal/discovery"
)

// rebuildSelectionList is a shared helper for regenerating a selection-based list.
// It handles the common pattern of setting disabled state, marking it dirty,
// calling updateSizes, and restoring the previous selection if it still exists.
// The caller is responsible for creating the new list before calling this helper.
func (m *UIModel) rebuildSelectionList(
	items []string,
	prevSelection string,
	disabledFlag *bool,
	selection *string,
	dirtyFlag *bool,
) {
	*disabledFlag = len(items) == 0
	if *disabledFlag {
		*selection = ""
	}
	m.updateSizes()
	*dirtyFlag = true

	// Restore selection if it still exists in the new list
	// Note: We set *selection directly (e.g., m.selection.Model or m.selection.Agent)
	// instead of calling list.Select(). This is because bubbles/list v0.10.3's Select()
	// doesn't restore visual cursor position when the same item remains selected - it only
	// updates internal state. The visual cursor will jump due to library limitations,
	// but the logical selection state is preserved correctly for downstream use.
	if prevSelection != "" && !*disabledFlag {
		found := false
		for _, itemName := range items {
			if itemName == prevSelection {
				*selection = prevSelection
				found = true
				break
			}
		}
		if !found {
			*selection = ""
		}
	}
}

// handleModelSkip regenerates the model list based on harness selection
// Expands provider: prefixes and handles discover:active keyword
func (m UIModel) handleModelSkip() (UIModel, tea.Cmd) {
	models := m.selection.Harness.SupportedModels

	var warnings []string
	expandedModels := make([]string, 0, len(models))
	for _, model := range models {
		switch {
		case strings.HasPrefix(model, discovery.PrefixProvider):
			providerID := strings.TrimPrefix(model, discovery.PrefixProvider)
			providerModels := m.app.Registry.GetModelsForProvider(providerID)
			if len(providerModels) == 0 {
				warnings = append(warnings, fmt.Sprintf("no models found for provider: %s (registry may not be loaded)", providerID))
			} else {
				expandedModels = append(expandedModels, providerModels...)
			}
		case model == discovery.KeywordDiscoverActive:
			activeModels := m.app.Registry.GetActiveModels()
			if len(activeModels) == 0 {
				warnings = append(warnings, "no active models found (check provider API keys and ensure registry is loaded)")
			} else {
				expandedModels = append(expandedModels, activeModels...)
			}
		default:
			expandedModels = append(expandedModels, model)
		}
	}

	var cmd tea.Cmd
	if len(warnings) > 0 {
		cmd = func() tea.Msg {
			return warningMsg{err: fmt.Errorf("%s", strings.Join(warnings, "; "))}
		}
	}

	uniqueModels := make([]string, 0, len(expandedModels))
	seen := make(map[string]bool)
	for _, model := range expandedModels {
		if !seen[model] {
			seen[model] = true
			uniqueModels = append(uniqueModels, model)
		}
	}
	models = uniqueModels

	// Save current model selection before regenerating list
	var prevModel string
	if item, ok := m.modelList.SelectedItem().(modelItem); ok {
		prevModel = item.name
	}

	m.modelList = newModelList(models, m.currentTheme)

	m.rebuildSelectionList(
		models,
		prevModel,
		&m.modelColumnDisabled,
		&m.selection.Model,
		&m.dirtyModel,
	)

	return m, cmd
}

// handleAgentSkip regenerates the agent list based on harness selection
// Preserves previous agent selection if still available
func (m UIModel) handleAgentSkip() (UIModel, tea.Cmd) {
	agents := m.selection.Harness.SupportedAgents

	var prevAgent string
	if item, ok := m.agentList.SelectedItem().(agentItem); ok {
		prevAgent = item.name
	}

	m.agentList = newAgentList(agents, m.currentTheme)

	m.rebuildSelectionList(
		agents,
		prevAgent,
		&m.agentColumnDisabled,
		&m.selection.Agent,
		&m.dirtyAgent,
	)

	return m, nil
}
