package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/megatherium/blunderbust/internal/domain"
)

func TestTemplateFilePickerFlow(t *testing.T) {
	// Create a temporary template file
	tempDir := t.TempDir()
	templatePath := filepath.Join(tempDir, "test.tmpl")
	templateContent := "echo hello {{.TicketID}}"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	app := newTestApp()
	m := NewUIModel(app, []domain.Harness{{Name: "test-harness"}})
	m.state = ViewStateConfirm
	m.selection = domain.Selection{
		Harness: domain.Harness{Name: "test-harness"},
		Ticket:  domain.Ticket{ID: "BB-123"},
	}

	// 1. Press 'C' to open filepicker for template
	cKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}}
	model, cmd, handled := m.handleKeyMsg(cKey)
	if !handled {
		t.Fatal("Expected 'C' key to be handled in ViewStateConfirm")
	}
	m = model.(UIModel)

	if m.state != ViewStateFilePicker {
		t.Errorf("Expected state to be ViewStateFilePicker, got %v", m.state)
	}
	if m.filePickerPurpose != fpPurposeTemplate {
		t.Errorf("Expected purpose to be fpPurposeTemplate, got %v", m.filePickerPurpose)
	}
	if cmd == nil {
		t.Error("Expected Init command when opening filepicker for template")
	}

	// 2. Simulate file selection in handleFilePickerKeyMsg
	// We'll directly call the command that would be returned by DidSelectFile
	loadCmd := m.loadTemplateFromFile(templatePath)
	msg := loadCmd()

	templateMsg, ok := msg.(templateLoadedMsg)
	if !ok {
		t.Fatalf("Expected templateLoadedMsg, got %T", msg)
	}
	if templateMsg.content != templateContent {
		t.Errorf("Expected content %q, got %q", templateContent, templateMsg.content)
	}

	// 3. Handle templateLoadedMsg - should enter inline edit mode
	model, cmd, handled = m.handleCoreMsgs(templateMsg)
	if !handled {
		t.Fatal("Expected templateLoadedMsg to be handled")
	}
	m = model.(UIModel)

	if m.state != ViewStateInlineEdit {
		t.Errorf("Expected state to be ViewStateInlineEdit, got %v", m.state)
	}
	if m.selection.Harness.CommandTemplate != "" {
		t.Errorf("Expected CommandTemplate to be empty (content goes to textarea), got %q", m.selection.Harness.CommandTemplate)
	}
	if m.inlineEditTextarea.Value() != templateContent {
		t.Errorf("Expected textarea to contain template content, got %q", m.inlineEditTextarea.Value())
	}
	if cmd != nil {
		t.Error("Expected nil command after template loaded")
	}

	// 4. Simulate Ctrl-y to accept the edit
	ctrlY := tea.KeyMsg{Type: tea.KeyCtrlY}
	model, _, handled = m.handleInlineEditKeyMsg(ctrlY)
	if !handled {
		t.Fatal("Expected Ctrl-y to be handled in inline edit mode")
	}
	m = model.(UIModel)

	if m.state != ViewStateConfirm {
		t.Errorf("Expected state to return to ViewStateConfirm after Ctrl-y, got %v", m.state)
	}
	if m.selection.Harness.CommandTemplate != templateContent {
		t.Errorf("Expected CommandTemplate to be %q after accept, got %q", templateContent, m.selection.Harness.CommandTemplate)
	}
}

func TestTemplateFilePicker_BinaryFile(t *testing.T) {
	// Create a temporary binary file
	tempDir := t.TempDir()
	binaryPath := filepath.Join(tempDir, "test.bin")
	binaryContent := []byte{0x00, 0xFF, 0x00, 0xFF}
	if err := os.WriteFile(binaryPath, binaryContent, 0644); err != nil {
		t.Fatalf("Failed to create binary file: %v", err)
	}

	app := newTestApp()
	m := NewUIModel(app, []domain.Harness{})

	loadCmd := m.loadTemplateFromFile(binaryPath)
	msg := loadCmd()

	testMsg, ok := msg.(templateErrorMsg)
	if !ok {
		t.Fatalf("Expected templateErrorMsg for binary file, got %T", msg)
	}
	if !strings.Contains(testMsg.err.Error(), "contains binary data") {
		t.Errorf("Expected binary data error, got: %v", testMsg.err)
	}

	// Should stay in FilePicker on templateErrorMsg
	model, _, handled := m.handleCoreMsgs(testMsg)
	if !handled {
		t.Fatal("Expected templateErrorMsg to be handled")
	}
	m = model.(UIModel)
	if m.state != ViewStateFilePicker {
		t.Errorf("Expected state to stay ViewStateFilePicker, got %v", m.state)
	}
	if len(m.warnings) == 0 {
		t.Error("Expected warning to be added on templateErrorMsg")
	}
}

func TestTemplateFilePicker_EscBackToConfirm(t *testing.T) {
	app := newTestApp()
	m := NewUIModel(app, []domain.Harness{})
	m.state = ViewStateFilePicker
	m.filePickerPurpose = fpPurposeTemplate

	escKey := tea.KeyMsg{Type: tea.KeyEscape}
	model, _, handled := m.handleFilePickerKeyMsg(escKey)
	if !handled {
		t.Fatal("Expected Esc to be handled in filepicker")
	}
	m = model.(UIModel)

	if m.state != ViewStateConfirm {
		t.Errorf("Expected state to be ViewStateConfirm after Esc, got %v", m.state)
	}
}

func TestTemplateFilePicker_KeyBindingsAndEditing(t *testing.T) {
	app := newTestApp()
	m := NewUIModel(app, []domain.Harness{})
	m.state = ViewStateFilePicker
	m.filePickerPurpose = fpPurposeTemplate

	// Test Ctrl-a (Toggle all extensions)
	ctrlA := tea.KeyMsg{Type: tea.KeyCtrlA}
	model, _, handled := m.handleFilePickerKeyMsg(ctrlA)
	if !handled {
		t.Fatal("Expected Ctrl-a to be handled")
	}
	m = model.(UIModel)
	if !m.filepicker.ShowAllExts {
		t.Error("Expected ShowAllExts to be true after Ctrl-a")
	}

	// Enable recents for swap testing
	m.filepicker.ShowRecents = true
	m.filepicker.Recents = []string{"/recent/1", "/recent/2"}

	// Test Tab (Swap View)
	tabKey := tea.KeyMsg{Type: tea.KeyTab}
	// bubbles/key matches string via msg.String(), so tea.KeyMsg{Type: tea.KeyTab} stringifies to "tab"
	model, _, handled = m.handleFilePickerKeyMsg(tabKey)
	if !handled {
		t.Fatal("Expected Tab to be handled")
	}
	m = model.(UIModel)
	// We can't directly check `recentFocus` because it's not exported,
	// but we know it's handled. We could test behavior if we needed.

	// Test 'l' (Edit CWD)
	lKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}}
	model, _, handled = m.handleFilePickerKeyMsg(lKey)
	if !handled {
		t.Fatal("Expected 'l' to be handled")
	}
	m = model.(UIModel)
	if !m.filepicker.EditingCwd {
		t.Fatal("Expected filepicker to enter EditingCwd state")
	}

	// Test Enter with invalid path (shows error)
	// First simulate typing an invalid path
	m.filepicker, _ = m.filepicker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b', 'a', 'd'}})
	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	model, _, handled = m.handleFilePickerKeyMsg(enterKey)
	if !handled {
		t.Fatal("Expected Enter to be handled while editing")
	}
	m = model.(UIModel)
	if m.filepicker.CwdError == "" {
		t.Error("Expected CwdError for invalid path")
	}
	if !m.filepicker.EditingCwd {
		t.Error("Expected to stay in EditingCwd when path is invalid")
	}

	// Test Esc cancels editing
	escKey := tea.KeyMsg{Type: tea.KeyEscape}
	model, _, handled = m.handleFilePickerKeyMsg(escKey)
	if !handled {
		t.Fatal("Expected Esc to be handled while editing")
	}
	m = model.(UIModel)
	if m.filepicker.EditingCwd {
		t.Error("Expected EditingCwd to be false after Esc")
	}
	if m.filepicker.CwdError != "" {
		t.Error("Expected CwdError to be cleared after Esc")
	}
}

func TestInlineEdit_EKeyEntersEditMode(t *testing.T) {
	app := newTestApp()
	m := NewUIModel(app, []domain.Harness{{Name: "test-harness"}})
	m.state = ViewStateConfirm
	m.selection = domain.Selection{
		Harness: domain.Harness{
			Name:            "test-harness",
			CommandTemplate: "echo hello {{.TicketID}}",
		},
		Ticket: domain.Ticket{ID: "BB-123"},
	}

	eKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	model, _, handled := m.handleKeyMsg(eKey)
	if !handled {
		t.Fatal("Expected 'e' key to be handled in ViewStateConfirm")
	}
	m = model.(UIModel)

	if m.state != ViewStateInlineEdit {
		t.Errorf("Expected state to be ViewStateInlineEdit, got %v", m.state)
	}
	if m.inlineEditMode != editModeCommand {
		t.Errorf("Expected editModeCommand, got %v", m.inlineEditMode)
	}
	if m.inlineEditTextarea.Value() != "echo hello {{.TicketID}}" {
		t.Errorf("Expected textarea to contain template, got %q", m.inlineEditTextarea.Value())
	}

	// Test Esc cancels and returns to confirm
	escKey := tea.KeyMsg{Type: tea.KeyEscape}
	model, _, handled = m.handleInlineEditKeyMsg(escKey)
	if !handled {
		t.Fatal("Expected Esc to be handled in inline edit mode")
	}
	m = model.(UIModel)

	if m.state != ViewStateConfirm {
		t.Errorf("Expected state to return to ViewStateConfirm, got %v", m.state)
	}
	if m.selection.Harness.CommandTemplate != "echo hello {{.TicketID}}" {
		t.Errorf("Expected CommandTemplate to be unchanged after cancel, got %q", m.selection.Harness.CommandTemplate)
	}
}

func TestInlineEdit_EKeyWithPrompt(t *testing.T) {
	app := newTestApp()
	m := NewUIModel(app, []domain.Harness{{Name: "test-harness"}})
	m.state = ViewStateConfirm
	m.selection = domain.Selection{
		Harness: domain.Harness{
			Name:           "test-harness",
			PromptTemplate: "Please help with {{.TicketID}}",
		},
		Ticket: domain.Ticket{ID: "BB-123"},
		Agent:  "codex",
	}

	eKey := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}}
	model, _, handled := m.handleKeyMsg(eKey)
	if !handled {
		t.Fatal("Expected 'e' key to be handled when agent is selected")
	}
	m = model.(UIModel)

	if m.state != ViewStateInlineEdit {
		t.Errorf("Expected state to be ViewStateInlineEdit, got %v", m.state)
	}
	if m.inlineEditMode != editModePrompt {
		t.Errorf("Expected editModePrompt when agent is selected, got %v", m.inlineEditMode)
	}
	if m.inlineEditTextarea.Value() != "Please help with {{.TicketID}}" {
		t.Errorf("Expected textarea to contain prompt template, got %q", m.inlineEditTextarea.Value())
	}
}
