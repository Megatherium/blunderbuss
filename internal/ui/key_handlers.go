package ui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/megatherium/blunderbust/internal/data/dolt"
)

func (m UIModel) handleModalKeyMsg() (tea.Model, tea.Cmd, bool) {
	if m.showModal {
		m.showModal = false
		return m, nil, true
	}
	return m, nil, false
}

func (m UIModel) handleQuitKeyMsg() (tea.Model, tea.Cmd, bool) {
	if m.state == ViewStateAgentOutput {
		m.viewingAgentID = ""
		m.state = ViewStateMatrix
		return m, nil, true
	}
	return m, tea.Quit, true
}

func (m UIModel) handleRefreshKeyMsg() (tea.Model, tea.Cmd, bool) {
	if m.state == ViewStateMatrix && m.focus == FocusTickets {
		m.state = ViewStateLoading
		return m, tea.Batch(
			loadTicketsCmd(m.app.Project().Store()),
			discoverWorktreesCmd(m.app),
			m.reloadTemplates(), // Also reload templates on refresh
		), true
	}
	return m, nil, false
}

func (m UIModel) handleBackKeyMsg() (tea.Model, tea.Cmd, bool) {
	if m.state == ViewStateConfirm {
		m.state = ViewStateMatrix
		return m, nil, true
	}
	if m.state == ViewStateAgentOutput {
		m.viewingAgentID = ""
		m.state = ViewStateMatrix
		return m, nil, true
	}
	if m.state == ViewStateMatrix && m.focus > FocusTickets {
		m.focus--
		return m, nil, true
	}
	return m, nil, false
}

func (m UIModel) handleInfoKeyMsg() (tea.Model, tea.Cmd, bool) {
	if m.state == ViewStateMatrix && m.focus == FocusTickets {
		if i, ok := m.ticketList.SelectedItem().(ticketItem); ok {
			m.showModal = true
			m.modalContent = "Loading bd show..."
			return m, loadModalCmd(i.ticket.ID), true
		}
	}
	return m, nil, false
}

func (m UIModel) handleToggleSidebarKeyMsg() (tea.Model, tea.Cmd, bool) {
	m.showSidebar = !m.showSidebar
	m.updateSizes()
	return m, nil, true
}

func (m UIModel) handleToggleThemeKeyMsg() (tea.Model, tea.Cmd, bool) {
	m.animState.nextTheme()
	m.currentTheme = m.animState.getCurrentTheme()
	m.ticketList.SetDelegate(newGradientDelegate(m.currentTheme))
	m.harnessList.SetDelegate(newGradientDelegate(m.currentTheme))
	m.modelList.SetDelegate(newGradientDelegate(m.currentTheme))
	m.agentList.SetDelegate(newGradientDelegate(m.currentTheme))
	m.dirtyTicket = true
	m.dirtyHarness = true
	m.dirtyModel = true
	m.dirtyAgent = true
	return m, nil, true
}

func (m UIModel) handleZoomKeyMsg() (tea.Model, tea.Cmd, bool) {
	// Toggle zoom mode
	m.ticketZoomEnabled = !m.ticketZoomEnabled

	// Update ticket delegate description lines
	if m.ticketDel != nil {
		if m.ticketZoomEnabled {
			m.ticketDel.SetDescLines(3) // 1 line for status/priority + 2 lines of description
		} else {
			m.ticketDel.SetDescLines(1) // Just status/priority
		}
	}

	// Recalculate layout with new zoom state
	m.layout = Compute(m.layout.TermWidth, m.layout.TermHeight, m.showSidebar, m.ticketZoomEnabled)
	m.updateSizes()

	// Mark all columns dirty since widths changed
	m.dirtyTicket = true
	m.dirtyHarness = true
	m.dirtyModel = true
	m.dirtyAgent = true

	return m, nil, true
}

func (m UIModel) handleFilePickerKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if m.state != ViewStateFilePicker {
		return m, nil, false
	}
	switch msg.String() {
	case "a":
		if m.filePickerPurpose == fpPurposeAddProject {
			currentDir := m.filepicker.CurrentDirectory
			if currentDir != "" {
				return m, m.checkAndPromptAddProject(currentDir), true
			}
			return m, nil, true
		}
	case "esc":
		if m.filepicker.EditingCwd {
			// Let the filepicker handle Esc to exit edit mode
			break
		}
		if m.filePickerPurpose == fpPurposeTemplate {
			m.state = ViewStateConfirm
		} else {
			m.state = ViewStateMatrix
		}
		return m, nil, true
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if m.filePickerPurpose == fpPurposeTemplate {
		didSelect, path := m.filepicker.DidSelectFile(msg)
		if didSelect {
			return m, m.loadTemplateFromFile(path), true
		}
	}

	return m, cmd, true
}

func (m UIModel) handleAddProjectModalKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if m.state != ViewStateAddProjectModal {
		return m, nil, false
	}
	switch msg.String() {
	case "y", "Y":
		return m, func() tea.Msg {
			return addProjectConfirmedMsg{path: m.pendingProjectPath}
		}, true
	case "n", "N", "q", "esc":
		return m, func() tea.Msg {
			return addProjectCancelledMsg{}
		}, true
	}
	return m, nil, true
}

func (m UIModel) handleErrorStateKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if m.state != ViewStateError {
		return m, nil, false
	}
	switch msg.String() {
	case "q", "Q":
		return m, tea.Quit, true
	case "r", "R":
		if m.retryStore != nil {
			m.state = ViewStateLoading
			return m, loadTicketsCmd(m.retryStore), true
		}
	case "s", "S":
		if m.retryStore != nil {
			if doltStore, ok := m.retryStore.(*dolt.Store); ok {
				if doltStore.CanRetryConnection() {
					m.state = ViewStateLoading
					return m, startServerAndRetryCmd(m.app, doltStore), true
				}
			}
		}
	}
	return m, nil, true
}

func (m UIModel) handleGlobalKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	// Don't process global keys when a list is in filtering mode
	if isFocusedListFiltering(m) {
		return m, nil, false
	}

	if key.Matches(msg, m.keys.Quit) {
		return m.handleQuitKeyMsg()
	}

	if key.Matches(msg, m.keys.Refresh) {
		if model, cmd, handled := m.handleRefreshKeyMsg(); handled {
			return model, cmd, true
		}
	}

	if key.Matches(msg, m.keys.Back) {
		if model, cmd, handled := m.handleBackKeyMsg(); handled {
			return model, cmd, true
		}
	}

	if key.Matches(msg, m.keys.Info) {
		if model, cmd, handled := m.handleInfoKeyMsg(); handled {
			return model, cmd, true
		}
	}

	if key.Matches(msg, m.keys.ToggleSidebar) {
		return m.handleToggleSidebarKeyMsg()
	}

	if key.Matches(msg, m.keys.ToggleTheme) {
		return m.handleToggleThemeKeyMsg()
	}

	if key.Matches(msg, m.keys.Zoom) {
		// Only enable zoom when ticket column is focused
		if m.focus == FocusTickets {
			return m.handleZoomKeyMsg()
		}
	}

	if key.Matches(msg, m.keys.PickTemplate) {
		if m.state == ViewStateConfirm {
			m.state = ViewStateFilePicker
			m.filePickerPurpose = fpPurposeTemplate
			m.filepicker.AllowedTypes = []string{".md", ".txt", ".tpl", ".tmpl"}
			m.filepicker.DirAllowed = true
			m.filepicker.FileAllowed = true
			
			// Set the initial directory to the project worktree if available, else current dir
			dir := m.selectedWorktree
			if dir == "" {
				dir = m.filepicker.CurrentDirectory
			}
			m.filepicker.CurrentDirectory = dir
			
			return m, m.filepicker.Init(), true
		}
	}

	return m, nil, false
}
