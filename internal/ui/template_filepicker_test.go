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

	// 3. Handle templateLoadedMsg
	model, cmd, handled = m.handleCoreMsgs(templateMsg)
	if !handled {
		t.Fatal("Expected templateLoadedMsg to be handled")
	}
	m = model.(UIModel)

	if m.state != ViewStateConfirm {
		t.Errorf("Expected state to return to ViewStateConfirm, got %v", m.state)
	}
	if m.selection.Harness.CommandTemplate != templateContent {
		t.Errorf("Expected CommandTemplate to be %q, got %q", templateContent, m.selection.Harness.CommandTemplate)
	}
	if cmd != nil {
		t.Error("Expected nil command after template loaded")
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
