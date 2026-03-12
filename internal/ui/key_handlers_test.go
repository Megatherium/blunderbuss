package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
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
