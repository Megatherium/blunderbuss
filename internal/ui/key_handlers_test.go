package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/megatherium/blunderbust/internal/data"
	"github.com/megatherium/blunderbust/internal/domain"
)

func TestHandleQuitKeyMsg_ExitsFromAgentOutput(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateAgentOutput
	model.viewingAgentID = "test-agent"

	newModel, cmd, handled := model.handleQuitKeyMsg()

	if !handled {
		t.Error("Expected message to be handled")
	}
	if newModel.(UIModel).viewingAgentID != "" {
		t.Error("Expected viewingAgentID to be cleared")
	}
	if newModel.(UIModel).state != ViewStateMatrix {
		t.Errorf("Expected state ViewStateMatrix, got %v", newModel.(UIModel).state)
	}
	if cmd != nil {
		t.Error("Expected no command")
	}
}

func TestHandleQuitKeyMsg_QuitsFromMatrix(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateMatrix

	_, cmd, handled := model.handleQuitKeyMsg()

	if !handled {
		t.Error("Expected message to be handled")
	}
	if cmd == nil {
		t.Error("Expected tea.Quit command")
	}
}

func TestHandleRefreshKeyMsg_IgnoresNonTicketFocus(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateMatrix
	model.focus = FocusHarness

	_, _, handled := model.handleRefreshKeyMsg()

	if handled {
		t.Error("Expected message not to be handled when not focused on tickets")
	}
}

func TestHandleBackKeyMsg_ExitsConfirmState(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateConfirm

	newModel, _, handled := model.handleBackKeyMsg()

	if !handled {
		t.Error("Expected message to be handled")
	}
	if newModel.(UIModel).state != ViewStateMatrix {
		t.Errorf("Expected state ViewStateMatrix, got %v", newModel.(UIModel).state)
	}
}

func TestHandleBackKeyMsg_ExitsAgentOutput(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateAgentOutput
	model.viewingAgentID = "test-agent"

	newModel, _, handled := model.handleBackKeyMsg()

	if !handled {
		t.Error("Expected message to be handled")
	}
	if newModel.(UIModel).viewingAgentID != "" {
		t.Error("Expected viewingAgentID to be cleared")
	}
	if newModel.(UIModel).state != ViewStateMatrix {
		t.Errorf("Expected state ViewStateMatrix, got %v", newModel.(UIModel).state)
	}
}

func TestHandleBackKeyMsg_RetreatsFocus(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateMatrix
	model.focus = FocusHarness

	newModel, _, handled := model.handleBackKeyMsg()

	if !handled {
		t.Error("Expected message to be handled")
	}
	if newModel.(UIModel).focus != FocusTickets {
		t.Errorf("Expected focus FocusTickets, got %v", newModel.(UIModel).focus)
	}
}

func TestHandleToggleSidebarKeyMsg_TogglesSidebar(t *testing.T) {
	model := NewTestModel()
	model.showSidebar = true

	newModel, _, handled := model.handleToggleSidebarKeyMsg()

	if !handled {
		t.Error("Expected message to be handled")
	}
	if newModel.(UIModel).showSidebar {
		t.Error("Expected sidebar to be hidden")
	}
}

func TestHandleToggleThemeKeyMsg_TogglesTheme(t *testing.T) {
	model := NewTestModel()
	initialTheme := model.currentTheme

	newModel, _, handled := model.handleToggleThemeKeyMsg()

	if !handled {
		t.Error("Expected message to be handled")
	}
	if newModel.(UIModel).currentTheme == initialTheme {
		t.Error("Expected theme to change")
	}
	if !newModel.(UIModel).dirtyTicket {
		t.Error("Expected dirtyTicket to be set")
	}
}

func TestHandleFilePickerKeyMsg_AcceptsProject(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateFilePicker
	model.filepicker.CurrentDirectory = "/test/project"

	_, cmd, handled := model.handleFilePickerKeyMsg(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
	)

	if !handled {
		t.Error("Expected message to be handled")
	}
	if cmd == nil {
		t.Error("Expected command to check project")
	}
}

func TestHandleFilePickerKeyMsg_Escapes(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateFilePicker

	newModel, _, handled := model.handleFilePickerKeyMsg(
		tea.KeyMsg{Type: tea.KeyEsc},
	)

	if !handled {
		t.Error("Expected message to be handled")
	}
	if newModel.(UIModel).state != ViewStateMatrix {
		t.Errorf("Expected state ViewStateMatrix, got %v", newModel.(UIModel).state)
	}
}

func TestHandleAddProjectModalKeyMsg_Confirms(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateAddProjectModal
	model.pendingProjectPath = "/test/project"

	_, cmd, handled := model.handleAddProjectModalKeyMsg(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}},
	)

	if !handled {
		t.Error("Expected message to be handled")
	}
	if cmd == nil {
		t.Error("Expected command to confirm")
	}
}

func TestHandleAddProjectModalKeyMsg_Cancels(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateAddProjectModal

	_, cmd, handled := model.handleAddProjectModalKeyMsg(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}},
	)

	if !handled {
		t.Error("Expected message to be handled")
	}
	if cmd == nil {
		t.Error("Expected command to cancel")
	}
}

func TestHandleErrorStateKeyMsg_Quits(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateError

	_, cmd, handled := model.handleErrorStateKeyMsg(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
	)

	if !handled {
		t.Error("Expected message to be handled")
	}
	if cmd == nil {
		t.Error("Expected tea.Quit command")
	}
}

func TestHandleErrorStateKeyMsg_Retries(t *testing.T) {
	model := NewTestModel()
	model.state = ViewStateError

	_, _, handled := model.handleErrorStateKeyMsg(
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}},
	)

	if !handled {
		t.Error("Expected message to be handled")
	}
}

func TestHandleGlobalKeyMsg_WhenFiltering_BlocksAllGlobalKeys(t *testing.T) {
	// Test that all global keys are blocked when ticket list is filtering
	globalKeys := []struct {
		name string
		key  tea.KeyMsg
	}{
		{"Quit-q", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}},
		{"Quit-ctrl+c", tea.KeyMsg{Type: tea.KeyCtrlC}},
		{"Refresh-r", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}},
		{"Back-esc", tea.KeyMsg{Type: tea.KeyEsc}},
		{"Info-i", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}},
		{"ToggleSidebar-p", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}},
		{"ToggleTheme-t", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}}},
		{"Zoom-z", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}},
	}

	// Helper to create a properly initialized model for testing
	setupModelWithLists := func() *UIModel {
		myApp := newTestApp()
		myApp.ActiveProject = "."
		if myApp.Stores == nil {
			myApp.Stores = make(map[string]data.TicketStore)
		}
		myApp.Stores["."] = &mockStore{}

		m := NewUIModel(myApp, nil)
		return &m
	}

	// Test Tickets column filtering
	for _, tc := range globalKeys {
		t.Run("Tickets_"+tc.name, func(t *testing.T) {
			model := setupModelWithLists()
			model.state = ViewStateMatrix
			model.focus = FocusTickets
			model.ticketList.SetFilterState(1) // 1 = Filtering state
			require.Equal(t, 1, int(model.ticketList.FilterState()), "list should be in Filtering state")

			_, _, handled := model.handleGlobalKeyMsg(tc.key)
			assert.False(t, handled, "global key %s should NOT be handled when ticket list is filtering", tc.name)
		})
	}

	// Test Harness column filtering
	for _, tc := range globalKeys {
		t.Run("Harness_"+tc.name, func(t *testing.T) {
			model := setupModelWithLists()
			model.state = ViewStateMatrix
			model.focus = FocusHarness
			model.harnessList.SetFilterState(1)
			require.Equal(t, 1, int(model.harnessList.FilterState()), "list should be in Filtering state")

			_, _, handled := model.handleGlobalKeyMsg(tc.key)
			assert.False(t, handled, "global key %s should NOT be handled when harness list is filtering", tc.name)
		})
	}

	// Test Model column filtering
	for _, tc := range globalKeys {
		t.Run("Model_"+tc.name, func(t *testing.T) {
			model := setupModelWithLists()
			model.state = ViewStateMatrix
			model.focus = FocusModel
			model.modelList.SetFilterState(1)
			require.Equal(t, 1, int(model.modelList.FilterState()), "list should be in Filtering state")

			_, _, handled := model.handleGlobalKeyMsg(tc.key)
			assert.False(t, handled, "global key %s should NOT be handled when model list is filtering", tc.name)
		})
	}

	// Test Agent column filtering
	for _, tc := range globalKeys {
		t.Run("Agent_"+tc.name, func(t *testing.T) {
			model := setupModelWithLists()
			model.state = ViewStateMatrix
			model.focus = FocusAgent
			model.agentList.SetFilterState(1)
			require.Equal(t, 1, int(model.agentList.FilterState()), "list should be in Filtering state")

			_, _, handled := model.handleGlobalKeyMsg(tc.key)
			assert.False(t, handled, "global key %s should NOT be handled when agent list is filtering", tc.name)
		})
	}
}

func TestHandleGlobalKeyMsg_WhenNotFiltering_ProcessesGlobalKeys(t *testing.T) {
	// Test that global keys ARE processed when not filtering
	myApp := newTestApp()
	myApp.ActiveProject = "."
	if myApp.Stores == nil {
		myApp.Stores = make(map[string]data.TicketStore)
	}
	myApp.Stores["."] = &mockStore{}

	model := NewUIModel(myApp, nil)
	model.state = ViewStateMatrix
	model.focus = FocusTickets
	// Don't set filter state, so not filtering

	// Test that 'p' toggles sidebar when not filtering
	_, _, handled := model.handleGlobalKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	assert.True(t, handled, "toggle sidebar should be handled when not filtering")
}

func TestHandleGlobalKeyMsg_NonMatrixState(t *testing.T) {
	// Verify that when not in Matrix state, global keys are still processed
	myApp := newTestApp()
	myApp.ActiveProject = "."
	if myApp.Stores == nil {
		myApp.Stores = make(map[string]data.TicketStore)
	}
	myApp.Stores["."] = &mockStore{}

	model := NewUIModel(myApp, nil)
	model.state = ViewStateAgentOutput
	model.viewingAgentID = "test-agent"

	// Even though not in Matrix state, quit key should exit agent output view
	newModel, cmd, handled := model.handleGlobalKeyMsg(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	assert.True(t, handled, "quit should be handled even when not in matrix state")
	assert.Nil(t, cmd, "should not return quit command in agent output state (returns to matrix instead)")
	assert.Equal(t, ViewStateMatrix, newModel.(UIModel).state, "should return to matrix state")
	assert.Empty(t, newModel.(UIModel).viewingAgentID, "should clear viewing agent ID")
}

func TestHandleFilePickerKeyMsg_FileSelectUpdatesRecents(t *testing.T) {
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "test.tmpl")
	if err := os.WriteFile(templatePath, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	app := newTestApp()
	model := NewUIModel(app, []domain.Harness{})
	model.state = ViewStateFilePicker
	model.filePickerPurpose = fpPurposeTemplate
	model.filepicker.FileAllowed = true
	model.filepicker.AllowedTypes = []string{".tmpl"}
	model.filepicker.CurrentDirectory = tempDir
	model.filepicker.Path = "/non-empty-path"
	model.filepicker.ShowAllExts = false

	initCmd := model.filepicker.Init()
	initMsg := initCmd()

	newFp, fpCmd := model.filepicker.Update(initMsg)
	model.filepicker = newFp
	if fpCmd != nil {
		fpCmd()
	}

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	newModel, cmd, handled := model.handleFilePickerKeyMsg(enterKey)
	updatedModel := newModel.(UIModel)

	if !handled {
		t.Fatal("Expected Enter to be handled in filepicker")
	}
	if cmd == nil {
		t.Fatal("Expected command to be returned when file is selected (should include RecentsChangedMsg)")
	}

	if len(updatedModel.filepicker.Recents) == 0 {
		t.Error("Expected Recents to be updated when file is selected")
	}

	if updatedModel.filepicker.Recents[0] != templatePath {
		t.Errorf("Expected selected file to be first in recents, got %s", updatedModel.filepicker.Recents[0])
	}
}
