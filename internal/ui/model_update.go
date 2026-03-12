package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/megatherium/blunderbust/internal/discovery"
	"github.com/megatherium/blunderbust/internal/domain"
)

func (m UIModel) handleNavigationKeysMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if m.state != ViewStateMatrix {
		return m, nil, false
	}

	// Don't process navigation keys when a list is in filtering mode
	if isFocusedListFiltering(m) {
		return m, nil, false
	}

	switch msg.String() {
	case "left", "h":
		return m.handleLeftNavigation()
	case "right", "l":
		return m.handleRightNavigation()
	case "tab":
		if m.focus < FocusAgent {
			m.advanceFocus()
		} else {
			m.focus = FocusSidebar
			m.sidebar.SetFocused(true)
		}
		return m, nil, true
	}
	return m, nil, false
}

func (m UIModel) handleLeftNavigation() (tea.Model, tea.Cmd, bool) {
	if m.focus == FocusSidebar {
		node := m.sidebar.State().CurrentNode()
		shouldCollapse := node != nil && len(node.Children) > 0 && node.IsExpanded
		if shouldCollapse {
			return m, nil, false // Let sidebar handle collapse
		}
	}
	if m.focus > FocusSidebar {
		m.retreatFocus()
		return m, nil, true
	}
	return m, nil, false
}

func (m UIModel) handleRightNavigation() (tea.Model, tea.Cmd, bool) {
	if m.focus == FocusSidebar {
		node := m.sidebar.State().CurrentNode()
		shouldExpand := node != nil && len(node.Children) > 0 && !node.IsExpanded
		if shouldExpand {
			return m, nil, false // Let sidebar handle expand
		}
	}
	if m.focus < FocusAgent {
		m.advanceFocus()
		return m, nil, true
	}
	return m, nil, false
}

func (m UIModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if model, cmd, handled := m.handleFilePickerKeyMsg(msg); handled {
		return model, cmd, handled
	}

	if model, cmd, handled := m.handleAddProjectModalKeyMsg(msg); handled {
		return model, cmd, handled
	}

	if model, cmd, handled := m.handleErrorStateKeyMsg(msg); handled {
		return model, cmd, handled
	}

	if model, cmd, handled := m.handleModalKeyMsg(); handled {
		return model, cmd, true
	}

	if model, cmd, handled := m.handleGlobalKeyMsg(msg); handled {
		return model, cmd, true
	}

	if model, cmd, handled := m.handleNavigationKeysMsg(msg); handled {
		return model, cmd, true
	}

	if key.Matches(msg, m.keys.Enter) {
		if m.focus == FocusSidebar {
			return m, nil, false
		}

		flashCmd := lockInCmd(m.focus)

		model, cmd := m.handleEnterKey()
		return model, tea.Batch(flashCmd, cmd), true
	}

	if model, cmd, handled := m.HandleSidebarAgentKeysMsg(msg); handled {
		return model, cmd, true
	}

	return m, nil, false
}

func (m UIModel) handleMatrixEnterKey() (tea.Model, tea.Cmd) {
	switch m.focus {
	case FocusSidebar:
		return m.handleSidebarEnterKey()
	case FocusTickets:
		return m.handleTicketsEnterKey()
	case FocusHarness:
		return m.handleHarnessEnterKey()
	case FocusModel:
		return m.handleModelEnterKey()
	case FocusAgent:
		return m.handleAgentEnterKey()
	}
	return m, nil
}

func (m UIModel) handleSidebarEnterKey() (tea.Model, tea.Cmd) {
	node := m.sidebar.State().CurrentNode()
	if node != nil && node.Type == domain.NodeTypeWorktree {
		m.selectedWorktree = node.Path
		m.sidebar.SetSelectedPath(node.Path)
		m.focus = FocusTickets
		m.sidebar.SetFocused(false)
		return m, nil
	}
	if node != nil && len(node.Children) > 0 {
		m.sidebar.State().ToggleExpand()
	}
	return m, nil
}

func (m UIModel) handleTicketsEnterKey() (tea.Model, tea.Cmd) {
	if i, ok := m.ticketList.SelectedItem().(ticketItem); ok {
		m.selection.Ticket = i.ticket

		if len(m.harnesses) == 1 {
			m.selection.Harness = m.harnesses[0]
			m, _ = m.handleModelSkip()
		}

		if m.focus < FocusAgent {
			m.advanceFocus()
		}
		return m, nil
	}
	return m, nil
}

func (m UIModel) handleHarnessEnterKey() (tea.Model, tea.Cmd) {
	if i, ok := m.harnessList.SelectedItem().(harnessItem); ok {
		m.selection.Harness = i.harness
		m, _ = m.handleModelSkip()
		m, _ = m.handleAgentSkip()
		if m.focus < FocusAgent {
			m.advanceFocus()
		}
		return m, nil
	}
	return m, nil
}

func (m UIModel) handleModelEnterKey() (tea.Model, tea.Cmd) {
	if i, ok := m.modelList.SelectedItem().(modelItem); ok {
		m.selection.Model = i.name
		m, _ = m.handleAgentSkip()
		if m.focus < FocusAgent {
			m.advanceFocus()
		}
		return m, nil
	}
	return m, nil
}

func (m UIModel) handleAgentEnterKey() (tea.Model, tea.Cmd) {
	if i, ok := m.agentList.SelectedItem().(agentItem); ok {
		m.selection.Agent = i.name
		m.state = ViewStateConfirm
		return m, nil
	}
	return m, nil
}

func (m UIModel) handleEnterKey() (tea.Model, tea.Cmd) {
	// Exit agent output view when Enter is pressed
	if m.state == ViewStateAgentOutput {
		m.viewingAgentID = ""
		m.state = ViewStateMatrix
		return m, nil
	}

	switch m.state {
	case ViewStateMatrix:
		return m.handleMatrixEnterKey()
	case ViewStateConfirm:
		m.state = ViewStateMatrix
		return m, m.launchCmd()
	}
	return m, nil
}

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

	m.modelColumnDisabled = len(models) == 0
	if m.modelColumnDisabled {
		m.selection.Model = ""
	}
	m.modelList = newModelList(models, m.currentTheme)
	m.updateSizes()
	m.dirtyModel = true

	// Restore model selection if it still exists in the new list
	// Note: We only set m.selection.Model here, not call m.modelList.Select().
	// This is because bubbles/list v0.10.3's Select() doesn't restore visual cursor
	// position when the same item remains selected - it only updates internal state.
	// The visual cursor will jump due to library limitations, but the logical selection
	// state is preserved correctly for downstream use.
	if prevModel != "" && !m.modelColumnDisabled {
		found := false
		for _, modelName := range models {
			if modelName == prevModel {
				m.selection.Model = prevModel
				found = true
				break
			}
		}
		// Clear selection if previously selected model no longer exists
		if !found {
			m.selection.Model = ""
		}
	}

	return m, cmd
}

func (m UIModel) handleAgentSkip() (UIModel, tea.Cmd) {
	agents := m.selection.Harness.SupportedAgents

	// Save current agent selection before regenerating list
	var prevAgent string
	if item, ok := m.agentList.SelectedItem().(agentItem); ok {
		prevAgent = item.name
	}

	m.agentColumnDisabled = len(agents) == 0
	if m.agentColumnDisabled {
		m.selection.Agent = ""
	}

	m.agentList = newAgentList(agents, m.currentTheme)
	m.updateSizes()
	m.dirtyAgent = true

	// Restore agent selection if it still exists in the new list
	// Note: We only set m.selection.Agent here, not call m.agentList.Select().
	// This is because bubbles/list v0.10.3's Select() doesn't restore visual cursor
	// position when the same item remains selected - it only updates internal state.
	// The visual cursor will jump due to library limitations, but the logical selection
	// state is preserved correctly for downstream use.
	if prevAgent != "" && !m.agentColumnDisabled {
		found := false
		for _, agentName := range agents {
			if agentName == prevAgent {
				m.selection.Agent = prevAgent
				found = true
				break
			}
		}
		// Clear selection if previously selected agent no longer exists
		if !found {
			m.selection.Agent = ""
		}
	}

	return m, nil
}

func (m *UIModel) updateKeyBindings() {
	switch m.state {
	case ViewStateMatrix:
		switch m.focus {
		case FocusSidebar:
			m.keys.Back.SetEnabled(false)
			m.keys.Refresh.SetEnabled(false)
			m.keys.Info.SetEnabled(false)
			m.keys.Enter.SetEnabled(true)
		case FocusTickets:
			m.keys.Back.SetEnabled(false)
			m.keys.Refresh.SetEnabled(true)
			m.keys.Info.SetEnabled(true)
			m.keys.Enter.SetEnabled(true)
		default:
			m.keys.Back.SetEnabled(true)
			m.keys.Refresh.SetEnabled(false)
			m.keys.Info.SetEnabled(false)
			m.keys.Enter.SetEnabled(true)
		}
		m.keys.ToggleSidebar.SetEnabled(true)
		m.keys.ToggleTheme.SetEnabled(true)
	case ViewStateError:
		m.keys.Back.SetEnabled(false)
		m.keys.Refresh.SetEnabled(false)
		m.keys.Enter.SetEnabled(false)
		m.keys.Info.SetEnabled(false)
		m.keys.ToggleSidebar.SetEnabled(false)
		m.keys.ToggleTheme.SetEnabled(false)
	default:
		m.keys.Back.SetEnabled(true)
		m.keys.Refresh.SetEnabled(false)
		m.keys.Enter.SetEnabled(true)
		m.keys.Info.SetEnabled(false)
		m.keys.ToggleSidebar.SetEnabled(false)
		m.keys.ToggleTheme.SetEnabled(true)
	}
}

func updateListCaches(m *UIModel) UIModel {
	if m.dirtyTicket || !m.initializedTicket {
		m.ticketViewCache = m.ticketList.View()
		m.dirtyTicket = false
		m.initializedTicket = true
	}
	if m.dirtyHarness || !m.initializedHarness {
		m.harnessViewCache = m.harnessList.View()
		m.dirtyHarness = false
		m.initializedHarness = true
	}
	if m.dirtyModel || !m.initializedModel {
		m.modelViewCache = m.modelList.View()
		m.dirtyModel = false
		m.initializedModel = true
	}
	if m.dirtyAgent || !m.initializedAgent {
		m.agentViewCache = m.agentList.View()
		m.dirtyAgent = false
		m.initializedAgent = true
	}
	return *m
}

func (m UIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Advance sidebar animation per event to ensure glitch effect runs
	// at a rate proportional to overall UI activity, matching old behavior.
	m.sidebar.TickAnimation()

	if m.state == ViewStateFilePicker {
		switch msg.(type) {
		case tea.KeyMsg, tea.WindowSizeMsg:
			// Let normal flow handle it so we process app-level keys and resize
		default:
			var fpCmd tea.Cmd
			m.filepicker, fpCmd = m.filepicker.Update(msg)
			if fpCmd != nil {
				return m, fpCmd
			}
		}
	}

	if newModel, cmd, handled := m.handleCoreMsgs(msg); handled {
		if uiModel, ok := newModel.(UIModel); ok {
			newModel = updateListCaches(&uiModel)
		}
		return newModel, cmd
	}
	if newModel, cmd, handled := m.handleProjectMsgs(msg); handled {
		if uiModel, ok := newModel.(UIModel); ok {
			newModel = updateListCaches(&uiModel)
		}
		return newModel, cmd
	}
	if newModel, cmd, handled := m.handleAgentMsgs(msg); handled {
		if uiModel, ok := newModel.(UIModel); ok {
			newModel = updateListCaches(&uiModel)
		}
		return newModel, cmd
	}

	uiModel, cmd := m.handleFocusUpdate(msg)
	uiModel.updateKeyBindings()
	newModel := updateListCaches(&uiModel)
	return newModel, cmd
}
