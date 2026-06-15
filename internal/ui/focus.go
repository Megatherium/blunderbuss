package ui

// FocusManager manages focus column navigation and dirty flag tracking
type FocusManager struct {
	model *UIModel
}

// newFocusManager creates a new FocusManager for the given UIModel
func newFocusManager(model *UIModel) *FocusManager {
	return &FocusManager{model: model}
}

// advanceFocus moves focus right, skipping disabled columns
func (fm *FocusManager) advance() {
	m := fm.model

	// Check if we can actually advance
	if m.focus >= FocusAgent {
		return
	}

	// Mark current focus column dirty
	fm.markColumnDirty(m.focus)

	// Find the next enabled column
	for nextFocus := m.focus + 1; nextFocus <= FocusAgent; nextFocus++ {
		// Skip disabled columns
		if nextFocus == FocusModel && m.modelColumnDisabled {
			continue
		}
		if nextFocus == FocusAgent && m.agentColumnDisabled {
			continue
		}
		// Found an enabled column, move to it
		if m.focus == FocusSidebar {
			m.sidebar.SetFocused(false)
		}
		m.focus = nextFocus
		// Mark new focus column dirty
		fm.markColumnDirty(m.focus)
		return
	}
	// No enabled column found, stay at current position
}

// retreatFocus moves focus left, skipping disabled columns
func (fm *FocusManager) retreat() {
	m := fm.model

	// Check if we can actually retreat
	if m.focus <= FocusSidebar {
		return
	}

	// Mark current focus column dirty
	fm.markColumnDirty(m.focus)

	// Find the previous enabled column
	for nextFocus := m.focus - 1; nextFocus >= FocusSidebar; nextFocus-- {
		// Skip disabled columns
		if nextFocus == FocusModel && m.modelColumnDisabled {
			continue
		}
		if nextFocus == FocusAgent && m.agentColumnDisabled {
			continue
		}
		// Found an enabled column, move to it
		if nextFocus == FocusSidebar {
			m.sidebar.SetFocused(true)
		}
		m.focus = nextFocus
		// Mark new focus column dirty
		fm.markColumnDirty(m.focus)
		return
	}
	// No enabled column found, stay at current position
}

// markColumnDirty sets the appropriate dirty flag based on the given focus type
func (fm *FocusManager) markColumnDirty(focus FocusColumn) {
	c := &fm.model.caches
	switch focus {
	case FocusTickets:
		c.dirtyTicket = true
	case FocusHarness:
		c.dirtyHarness = true
	case FocusModel:
		c.dirtyModel = true
	case FocusAgent:
		c.dirtyAgent = true
	}
}

// markAllColumnsDirty sets all column dirty flags to true
func (fm *FocusManager) markAllColumnsDirty() {
	c := &fm.model.caches
	c.dirtyTicket = true
	c.dirtyHarness = true
	c.dirtyModel = true
	c.dirtyAgent = true
}

// advanceFocus moves focus right, skipping disabled columns
func (m *UIModel) advanceFocus() {
	newFocusManager(m).advance()
}

// retreatFocus moves focus left, skipping disabled columns
func (m *UIModel) retreatFocus() {
	newFocusManager(m).retreat()
}

// markColumnDirty sets the appropriate dirty flag based on the given focus type
func (m *UIModel) markColumnDirty(f FocusColumn) {
	newFocusManager(m).markColumnDirty(f)
}

// markAllColumnsDirty sets all column dirty flags to true
func (m *UIModel) markAllColumnsDirty() {
	newFocusManager(m).markAllColumnsDirty()
}
