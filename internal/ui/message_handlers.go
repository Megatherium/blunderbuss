package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/megatherium/blunderbust/internal/discovery"
	"github.com/megatherium/blunderbust/internal/domain"
	"github.com/megatherium/blunderbust/internal/exec/tmux"
)

// Performance counters for debug instrumentation (--debug flag).
// These track the ticket refresh cycle to diagnose CPU burn issues.
var (
	perfTicketCheckCount   int64
	perfLastCheckTime      time.Time
	perfAutoRefreshCount   int64
	perfLastRefreshTime    time.Time
	perfTicketsLoadedCount int64
	perfLastLoadedTime     time.Time
)

func (m UIModel) handleTicketsLoaded(msg ticketsLoadedMsg) (tea.Model, tea.Cmd) {
	perfTicketsLoadedCount++
	now := time.Now()
	if m.app.Opts.Debug {
		sinceLast := now.Sub(perfLastLoadedTime)
		fmt.Fprintf(os.Stderr, "[DEBUG][perf] ticketsLoaded #%d lastLoadAgo=%v count=%d goroutines=%d\n",
			perfTicketsLoadedCount, sinceLast.Round(time.Millisecond), len(msg.tickets), runtime.NumGoroutine())
	}
	perfLastLoadedTime = now

	var prevTicketID string
	if i, ok := m.ticketList.SelectedItem().(ticketItem); ok {
		prevTicketID = i.ticket.ID
	}

	// Preserve filter state before recreating the list.
	// We save both the value and whether the user was actively typing (Filtering)
	// vs had finished typing (FilterApplied).
	var savedFilterValue string
	var savedFilterState list.FilterState
	var savedCursorPos int
	if m.ticketList.FilterState() != list.Unfiltered {
		savedFilterValue = m.ticketList.FilterValue()
		savedFilterState = m.ticketList.FilterState()
		savedCursorPos = m.ticketList.FilterInput.Position()
	}

	// Ensure we have a live ticketDelegate reference; create one if missing.
	if m.ticketDel == nil {
		m.ticketDel = newTicketDelegate(m.currentTheme)
	} else {
		m.ticketDel.applyTheme(m.currentTheme)
	}

	if len(msg.tickets) == 0 {
		if m.app.Project() == nil || m.app.Project().Store() == nil {
			m.ticketList = createErrorList("Couldn't load ticket list:\nStore initialization failed", m.currentTheme)
			m.sidebar.SetStoreError(true)
			if m.state == ViewStateLoading {
				m.state = ViewStateMatrix
			}
			return m, nil
		}
		items := []list.Item{emptyTicketItem{}}
		if m.ticketDel != nil {
			m.ticketDel.UpdateMaxTitleWidth(items)
		}
		m.ticketList = list.New(items, m.ticketDel, 0, 0)
		m.ticketList.SetShowStatusBar(false)
		m.sidebar.SetStoreError(false)
	} else {
		items := make([]list.Item, 0, len(msg.tickets))
		for i := range msg.tickets {
			items = append(items, ticketItem{ticket: msg.tickets[i], project: msg.project})
		}
		if m.ticketDel != nil {
			m.ticketDel.UpdateMaxTitleWidth(items)
		}
		m.ticketList = list.New(items, m.ticketDel, 0, 0)
		m.sidebar.SetStoreError(false)
	}
	initList(&m.ticketList, 0, 0, "Select a Ticket")
	if m.state == ViewStateLoading {
		m.state = ViewStateMatrix
	}
	m.updateSizes()
	m.dirtyTicket = true

	// Restore filter state if one was active.
	// SetFilterText applies filter and sets state to FilterApplied.
	// It also calls GoToStart() which resets selection to index 0.
	// SetFilterState restores Filtering state and focuses the input.
	// We restore cursor position and selection after.
	if savedFilterValue != "" {
		m.ticketList.SetFilterText(savedFilterValue)
		if savedFilterState == list.Filtering {
			m.ticketList.SetFilterState(list.Filtering)
		}
		m.ticketList.FilterInput.SetCursor(savedCursorPos)
	} else if savedFilterState == list.Filtering {
		// User activated the filter ("/") but hasn't typed anything yet.
		// Run filter with empty value so filteredItems contains all items,
		// then restore Filtering state to keep the input focused.
		m.ticketList.SetFilterText("")
		m.ticketList.SetFilterState(list.Filtering)
		m.ticketList.FilterInput.SetCursor(savedCursorPos)
	}

	// Restore selection in the filtered list.
	// When no filter is active, use unfiltered index.
	// When filter is active, find the ticket in VisibleItems() and select there.
	if prevTicketID != "" && savedFilterValue == "" {
		// No filter: use unfiltered index
		foundIndex := -1
		for idx, ticket := range msg.tickets {
			if ticket.ID == prevTicketID {
				m.selection.Ticket = ticket
				foundIndex = idx
				break
			}
		}
		if foundIndex >= 0 {
			m.ticketList.Select(foundIndex)
		} else {
			m.selection.Ticket = domain.Ticket{}
		}
	} else if prevTicketID != "" {
		// Filter is active: find selection in filtered items
		// and select by filtered index to keep selection across refreshes
		visibleItems := m.ticketList.VisibleItems()
		foundIdx := -1
		for idx, item := range visibleItems {
			if ti, ok := item.(ticketItem); ok && ti.ticket.ID == prevTicketID {
				m.selection.Ticket = ti.ticket
				foundIdx = idx
				break
			}
		}
		if foundIdx >= 0 {
			m.ticketList.Select(foundIdx)
		} else {
			// Ticket not in filtered results (e.g., filter excludes it)
			// Reset selection to first filtered item or clear
			m.selection.Ticket = domain.Ticket{}
			if len(visibleItems) > 0 {
				m.ticketList.Select(0)
			}
		}
	}

	return m, nil
}

func (m UIModel) handleErrMsg(msg errMsg) (tea.Model, tea.Cmd) {
	m.err = msg.err
	m.state = ViewStateError
	if msg.showRetryOptions {
		if project := m.app.Project(); project != nil {
			m.retryStore = project.Store()
		}
	} else {
		m.retryStore = nil
	}
	return m, nil
}

const maxWarnings = 50

func (m UIModel) handleWarningMsg(msg warningMsg) (tea.Model, tea.Cmd) {
	m.warnings = append(m.warnings, msg.err.Error())
	if len(m.warnings) > maxWarnings {
		m.warnings = m.warnings[len(m.warnings)-maxWarnings:]
	}
	if m.app != nil && m.app.Opts.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG][i29d] handleWarningMsg: warnings count=%d (latest: %s)\n", len(m.warnings), msg.err.Error())
	}
	return m, nil
}

func (m UIModel) handleInfoMsg(msg infoMsg) (tea.Model, tea.Cmd) {
	m.warnings = append(m.warnings, msg.message)
	if len(m.warnings) > maxWarnings {
		m.warnings = m.warnings[len(m.warnings)-maxWarnings:]
	}
	if m.app != nil && m.app.Opts.Debug {
		fmt.Fprintf(os.Stderr, "[DEBUG][i29d] handleInfoMsg: warnings count=%d (latest: %s)\n", len(m.warnings), msg.message)
	}
	return m, nil
}

func (m UIModel) handleLaunchResult(msg launchResultMsg) (tea.Model, tea.Cmd) {
	m.launchResult = msg.res
	m.err = msg.err

	if msg.err != nil {
		m.state = ViewStateError
		return m, nil
	}

	if msg.res != nil && msg.res.LauncherID != "" {
		selection := m.selection
		if msg.spec != nil {
			selection = msg.spec.Selection
		}

		agentID := msg.res.LauncherID
		agentInfo := &domain.AgentInfo{
			ID:           agentID,
			Name:         selection.Ticket.ID,
			LauncherID:   msg.res.LauncherID,
			WorktreePath: m.selectedWorktree,
			Status:       domain.AgentRunning,
			StartedAt:    time.Now(),
			TicketID:     selection.Ticket.ID,
			TicketTitle:  selection.Ticket.Title,
			HarnessName:  selection.Harness.Name,
			ModelName:    selection.Model,
			AgentName:    selection.Agent,
		}

		var capture *tmux.OutputCapture
		launcherID := msg.res.LauncherID
		if launcherID != "" && m.app.Runner() != nil && msg.res.LauncherType == domain.LauncherTypeTmux {
			capture = tmux.NewOutputCapture(m.app.Runner(), launcherID)
			path, captureErr := capture.Start(context.Background())
			if captureErr != nil {
				m.warnings = append(m.warnings, fmt.Sprintf("Failed to capture output: %v", captureErr))
				capture = nil
			}
			_ = path
		}

		m.agents[agentID] = &RunningAgent{
			Info:    agentInfo,
			Capture: capture,
		}

		AddAgentNodeToSidebar(&m, agentInfo)

		m.state = ViewStateMatrix

		return m, tea.Batch(
			pollAgentStatusCmd(m.app, agentID, msg.res.LauncherID),
			startAgentMonitoringCmd(agentID),
			saveRunningAgentCmd(m.app, msg.spec, msg.res, m.selectedWorktree),
		)
	}

	m.state = ViewStateMatrix
	return m, nil
}

func (m UIModel) handleWindowSizeMsg(msg tea.WindowSizeMsg) (UIModel, tea.Cmd) {
	m.layout = Compute(msg.Width, msg.Height, m.showSidebar, m.ticketZoomEnabled)
	m.updateSizes()
	m.dirtyTicket = true
	m.dirtyHarness = true
	m.dirtyModel = true
	m.dirtyAgent = true
	return m, nil
}

func (m UIModel) handleWorktreesDiscovered(msg worktreesDiscoveredMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.warnings = append(m.warnings, fmt.Sprintf("Worktree discovery: %v", msg.err))
		return m, nil
	}

	state := m.sidebar.State()
	prevNode := state.CurrentNode()
	prevSelectedPath := m.selectedWorktree

	m.sidebar, _ = m.sidebar.Update(SidebarNodesMsg{Nodes: msg.nodes})

	if prevSelectedPath != "" {
		found := false
		for _, info := range m.sidebar.State().FlatNodes {
			if info.Node.Path == prevSelectedPath {
				found = true
				break
			}
		}
		if found {
			m.selectedWorktree = prevSelectedPath
			m.sidebar.SetSelectedPath(prevSelectedPath)
		} else if len(msg.nodes) > 0 && len(msg.nodes[0].Children) > 0 {
			initialPath := msg.nodes[0].Children[0].Path
			m.selectedWorktree = initialPath
			m.sidebar.SetSelectedPath(initialPath)
		}
	} else if len(msg.nodes) > 0 && len(msg.nodes[0].Children) > 0 {
		initialPath := msg.nodes[0].Children[0].Path
		m.selectedWorktree = initialPath
		m.sidebar.SetSelectedPath(initialPath)
	}

	if prevNode != nil {
		for i, info := range m.sidebar.State().FlatNodes {
			if info.Node.Path == prevNode.Path {
				m.sidebar.State().Cursor = i
				break
			}
		}
	}

	for _, running := range m.agents {
		if running != nil && running.Info != nil {
			AddAgentNodeToSidebar(&m, running.Info)
		}
	}

	return m, nil
}

func (m UIModel) handleRunningAgentsLoaded(msg runningAgentsLoadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.warnings = append(m.warnings, fmt.Sprintf("Running agents load: %v", msg.err))
		return m, nil
	}

	var cmds []tea.Cmd
	for _, persisted := range msg.agents {
		agentID := PersistedAgentID(persisted)
		if existing, ok := m.agents[agentID]; ok && existing != nil {
			existing.Info.Status = domain.AgentRunning
			continue
		}

		info := &domain.AgentInfo{
			ID:           agentID,
			Name:         persisted.Ticket,
			LauncherID:   persisted.LauncherID,
			WorktreePath: persisted.WorktreePath,
			Status:       domain.AgentRunning,
			StartedAt:    persisted.StartedAt,
			TicketID:     persisted.Ticket,
			TicketTitle:  persisted.TicketTitle,
			HarnessName:  persisted.HarnessName,
			ModelName:    persisted.Model,
			AgentName:    persisted.Agent,
		}
		m.agents[agentID] = &RunningAgent{Info: info}
		AddAgentNodeToSidebar(&m, info)

		if persisted.LauncherID != "" {
			cmds = append(cmds,
				pollAgentStatusCmd(m.app, agentID, persisted.LauncherID),
				startAgentMonitoringCmd(agentID),
			)
		}
	}

	if len(cmds) == 0 {
		return m, nil
	}
	return m, tea.Batch(cmds...)
}

func (m UIModel) handleWorktreeSelected(msg WorktreeSelectedMsg) (tea.Model, tea.Cmd) {
	m.selectedWorktree = msg.Path
	m.sidebar.SetSelectedPath(msg.Path)

	m.focus = FocusTickets
	m.sidebar.SetFocused(false)
	m.dirtyTicket = true
	return m, nil
}

func (m UIModel) handleTicketUpdateCheck() (tea.Model, tea.Cmd) {
	perfTicketCheckCount++
	now := time.Now()
	if m.app.Opts.Debug {
		sinceLast := now.Sub(perfLastCheckTime)
		fmt.Fprintf(os.Stderr, "[DEBUG][perf] ticketUpdateCheck #%d lastCheckAgo=%v lastDBUpdate=%v goroutines=%d\n",
			perfTicketCheckCount, sinceLast.Round(time.Millisecond), m.lastTicketUpdate, runtime.NumGoroutine())
	}
	perfLastCheckTime = now

	if m.app.Project() == nil {
		return m, tea.Tick(ticketPollingInterval, func(t time.Time) tea.Msg {
			return ticketUpdateCheckMsg{}
		})
	}

	store := m.app.Project().Store()
	if store == nil {
		return m, tea.Tick(ticketPollingInterval, func(t time.Time) tea.Msg {
			return ticketUpdateCheckMsg{}
		})
	}
	return m, checkTicketUpdatesCmd(store, m.lastTicketUpdate, m.app.Opts.Debug)
}

func (m UIModel) handleTicketUpdateCheckNeeded() (tea.Model, tea.Cmd) {
	return m, tea.Tick(ticketPollingInterval, func(t time.Time) tea.Msg {
		return ticketUpdateCheckMsg{}
	})
}

func (m UIModel) handleTicketsAutoRefreshed(msg ticketsAutoRefreshedMsg) (tea.Model, tea.Cmd) {
	perfAutoRefreshCount++
	now := time.Now()
	if m.app.Opts.Debug {
		sinceLast := now.Sub(perfLastRefreshTime)
		fmt.Fprintf(os.Stderr, "[DEBUG][perf] ticketsAutoRefreshed #%d lastRefreshAgo=%v dbUpdatedAt=%v goroutines=%d\n",
			perfAutoRefreshCount, sinceLast.Round(time.Millisecond), msg.dbUpdatedAt, runtime.NumGoroutine())
	}
	perfLastRefreshTime = now

	if !msg.dbUpdatedAt.IsZero() {
		m.lastTicketUpdate = msg.dbUpdatedAt
	}
	m.refreshedRecently = true
	m.refreshAnimationFrame = 0

	cmds := []tea.Cmd{loadTicketsCmd(m.app.Project(), m.app.Opts.Debug), discoverWorktreesCmd(m.app)}

	if m.app.Fonts.HasNerdFont {
		cmds = append(cmds, tea.Tick(animationTickInterval, func(t time.Time) tea.Msg {
			return refreshAnimationTickMsg{}
		}))
	}

	cmds = append(cmds,
		tea.Tick(ticketPollingInterval, func(t time.Time) tea.Msg {
			return clearRefreshIndicatorMsg{}
		}),
		tea.Tick(ticketPollingInterval, func(t time.Time) tea.Msg {
			return ticketUpdateCheckMsg{}
		}))

	return m, tea.Batch(cmds...)
}

func (m UIModel) handleClearRefreshIndicator() (tea.Model, tea.Cmd) {
	m.refreshedRecently = false
	return m, nil
}

func (m UIModel) handleAddProjectConfirmed(msg addProjectConfirmedMsg) (tea.Model, tea.Cmd) {
	projectDir := msg.path

	for _, project := range m.app.GetProjects() {
		if filepath.Clean(project.Dir) == filepath.Clean(projectDir) {
			m.warnings = append(m.warnings, fmt.Sprintf("Project already exists: %s", projectDir))
			m.state = ViewStateFilePicker
			return m, nil
		}
	}

	project := domain.Project{
		Dir:  projectDir,
		Name: filepath.Base(projectDir),
	}
	m.app.AddProject(project)

	ctx := context.Background()
	beadsDir := filepath.Join(projectDir, ".beads")
	store, err := m.app.CreateStore(ctx, beadsDir)
	if err != nil {
		m.warnings = append(m.warnings, fmt.Sprintf("Failed to create store for %s: %v", projectDir, err))
		m.state = ViewStateFilePicker
		return m, nil
	}
	m.app.AddStore(projectDir, store)

	if err := m.app.SetActiveProject(ctx, projectDir); err != nil {
		m.warnings = append(m.warnings, fmt.Sprintf("Failed to activate project %s: %v", projectDir, err))
		m.state = ViewStateFilePicker
		return m, nil
	}

	if err := m.app.SaveConfig(); err != nil {
		m.warnings = append(m.warnings, fmt.Sprintf("Failed to save config: %v", err))
	}

	m.state = ViewStateMatrix
	m.pendingProjectPath = ""

	return m, tea.Batch(
		m.loadRegistryCmd(),
		m.continueInitAfterRegistry(),
		func() tea.Msg {
			return warningMsg{fmt.Errorf("added project: %s", projectDir)}
		},
	)
}

// handleTemplatesReloaded handles the successful reloading of templates.
// Updates the UI model with the new harness configurations containing fresh templates.
func (m UIModel) handleTemplatesReloaded(msg TemplatesReloadedMsg) (tea.Model, tea.Cmd) {
	// Update the harnesses with fresh templates
	m.harnesses = msg.Harnesses

	// Mark harness cache as dirty to force re-render with new templates
	m.dirtyHarness = true

	// If we're in the matrix view, update the harness list
	if m.state == ViewStateMatrix {
		// Rebuild the harness list with updated templates
		var registry *discovery.Registry
		if m.app != nil {
			registry = m.app.Registry
		}
		hl := newHarnessList(m.harnesses, registry, m.currentTheme)
		initList(&hl, 0, 0, "Select a Harness")
		m.harnessList = hl

		// Try to preserve the current selection if possible
		if len(m.harnesses) > 0 {
			if i, ok := m.harnessList.SelectedItem().(harnessItem); ok {
				m.selection.Harness = i.harness
				m, _ = m.handleModelSkip()
				m, _ = m.handleAgentSkip()
			}
		}
	}

	return m, func() tea.Msg {
		return infoMsg{message: "templates reloaded successfully"}
	}
}

// handleTemplateReloadError handles errors that occur during template reloading.
// Shows an error message but allows the application to continue running.
func (m UIModel) handleTemplateReloadError(msg TemplateReloadErrorMsg) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return errMsg{err: fmt.Errorf("failed to reload templates: %w", msg.Error), showRetryOptions: false}
	}
}
