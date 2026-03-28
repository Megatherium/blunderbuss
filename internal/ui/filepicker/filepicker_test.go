package filepicker

import (
	"io/fs"
	"os"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type mockDirEntry struct {
	name string
}

func (m mockDirEntry) Name() string               { return m.name }
func (m mockDirEntry) IsDir() bool                { return false }
func (m mockDirEntry) Type() fs.FileMode          { return 0 }
func (m mockDirEntry) Info() (fs.FileInfo, error) { return nil, nil }

func TestPruneRecents_RemovesDeadEntries(t *testing.T) {
	tmpDir := t.TempDir()
	validPath1 := tmpDir + "/file1.txt"
	validPath3 := tmpDir + "/file3.txt"
	if err := os.WriteFile(validPath1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if err := os.WriteFile(validPath3, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	m := New()
	m.Recents = []string{
		validPath1,
		"/dead/file2.txt",
		validPath3,
	}
	m.recentSelect = 1

	m.PruneRecents()

	if len(m.Recents) != 2 {
		t.Errorf("Expected 2 recents after pruning, got %d", len(m.Recents))
	}
	if m.Recents[0] != validPath1 {
		t.Errorf("Expected %s, got %s", validPath1, m.Recents[0])
	}
	if m.Recents[1] != validPath3 {
		t.Errorf("Expected %s, got %s", validPath3, m.Recents[1])
	}
	if m.recentSelect != 1 {
		t.Errorf("Expected recentSelect to be 1 (shifted from dead entry before), got %d", m.recentSelect)
	}
}

func TestPruneRecents_AdjustsCursorWhenOutOfBounds(t *testing.T) {
	tmpDir := t.TempDir()
	validPath := tmpDir + "/file1.txt"
	if err := os.WriteFile(validPath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	m := New()
	m.Recents = []string{validPath}
	m.recentSelect = 5

	m.PruneRecents()

	if m.recentSelect != 0 {
		t.Errorf("Expected recentSelect to be 0 after pruning to 1 item, got %d", m.recentSelect)
	}
}

func TestPruneRecents_EmptyListDoesNothing(t *testing.T) {
	m := New()
	m.Recents = []string{}
	m.recentSelect = 0

	m.PruneRecents()

	if len(m.Recents) != 0 {
		t.Errorf("Expected 0 recents, got %d", len(m.Recents))
	}
}

func TestDidSelectRecent_PrunesDeadEntryOnSelect(t *testing.T) {
	tmpDir := t.TempDir()
	validPath1 := tmpDir + "/file1.txt"
	validPath3 := tmpDir + "/file3.txt"
	if err := os.WriteFile(validPath1, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if err := os.WriteFile(validPath3, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	m := New()
	m.Recents = []string{
		validPath1,
		"/dead/file2.txt",
		validPath3,
	}
	m.recentSelect = 1
	m.FileAllowed = true

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	didSelect, path := m.didSelectRecent(enterKey)

	if didSelect {
		t.Error("Expected didSelect to be false for dead entry")
	}
	if path != "" {
		t.Errorf("Expected empty path, got %s", path)
	}
	if len(m.Recents) != 2 {
		t.Errorf("Expected 2 recents after pruning dead entry, got %d", len(m.Recents))
	}
	if m.recentSelect != 1 {
		t.Errorf("Expected recentSelect to be 1 after pruning, got %d", m.recentSelect)
	}
}

func TestDidSelectRecent_SelectsValidEntry(t *testing.T) {
	tmpDir := t.TempDir()
	validPath := tmpDir + "/valid.txt"
	if err := os.WriteFile(validPath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	m := New()
	m.Recents = []string{"/dead/file.txt", validPath}
	m.recentSelect = 1
	m.FileAllowed = true

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	didSelect, path := m.didSelectRecent(enterKey)

	if !didSelect {
		t.Error("Expected didSelect to be true for valid entry")
	}
	if path != validPath {
		t.Errorf("Expected %s, got %s", validPath, path)
	}
}

func TestDidSelectFile_PublicAPI_ReturnsUpdatedModel(t *testing.T) {
	tmpDir := t.TempDir()
	validPath := tmpDir + "/valid.txt"
	if err := os.WriteFile(validPath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	m := New()
	m.Recents = []string{"/dead/file.txt", validPath}
	m.recentSelect = 1
	m.FileAllowed = true
	m.recentFocus = true

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, didSelect, path := m.DidSelectFile(enterKey)

	if !didSelect {
		t.Error("Expected didSelect to be true for valid entry")
	}
	if path != validPath {
		t.Errorf("Expected %s, got %s", validPath, path)
	}
	if len(updatedModel.Recents) != 2 {
		t.Errorf("Expected 2 recents in returned model (no pruning for valid entry), got %d", len(updatedModel.Recents))
	}
}

func TestDidSelectFile_PublicAPI_PrunesDeadEntry(t *testing.T) {
	tmpDir := t.TempDir()
	validPath := tmpDir + "/valid.txt"
	if err := os.WriteFile(validPath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	m := New()
	m.Recents = []string{validPath, "/dead/file.txt"}
	m.recentSelect = 1
	m.FileAllowed = true
	m.recentFocus = true

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, didSelect, path := m.DidSelectFile(enterKey)

	if didSelect {
		t.Error("Expected didSelect to be false for dead entry")
	}
	if path != "" {
		t.Errorf("Expected empty path for dead entry, got %s", path)
	}
	if len(updatedModel.Recents) != 1 {
		t.Errorf("Expected 1 recent in returned model (dead entry pruned), got %d", len(updatedModel.Recents))
	}
	if updatedModel.Recents[0] != validPath {
		t.Errorf("Expected remaining recent to be validPath, got %s", updatedModel.Recents[0])
	}
}

func TestDidSelectRecent_SelectsSecondEntry(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := tmpDir + "/file1.txt"
	file2 := tmpDir + "/file2.txt"
	if err := os.WriteFile(file1, []byte("FILE1_CONTENT"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("FILE2_CONTENT"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	m := New()
	m.Recents = []string{file1, file2}
	m.recentSelect = 1
	m.recentFocus = true
	m.FileAllowed = true

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	didSelect, path := m.didSelectRecent(enterKey)

	if !didSelect {
		t.Error("Expected didSelect to be true")
	}
	if path != file2 {
		t.Errorf("Expected %s (2nd entry), got %s", file2, path)
	}
}

func TestDidSelectFile_SelectsSecondRecentEntry(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := tmpDir + "/file1.txt"
	file2 := tmpDir + "/file2.txt"
	if err := os.WriteFile(file1, []byte("FILE1_CONTENT"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("FILE2_CONTENT"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	m := New()
	m.Recents = []string{file1, file2}
	m.recentSelect = 1
	m.recentFocus = true
	m.FileAllowed = true

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}

	m, _ = m.Update(enterKey)
	_, didSelect, path := m.DidSelectFile(enterKey)

	if !didSelect {
		t.Error("Expected didSelect to be true")
	}
	if path != file2 {
		t.Errorf("Expected %s (2nd entry), got %s", file2, path)
	}
}

func TestDidSelectDisabledFile_PublicAPI_ReturnsModel(t *testing.T) {
	tmpDir := t.TempDir()
	validPath := tmpDir + "/valid.txt"
	if err := os.WriteFile(validPath, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	m := New()
	m.Recents = []string{"/dead/file.txt", validPath}
	m.recentSelect = 1
	m.FileAllowed = true
	m.recentFocus = true

	enterKey := tea.KeyMsg{Type: tea.KeyEnter}
	updatedModel, didSelect, path := m.DidSelectDisabledFile(enterKey)

	if didSelect {
		t.Error("Expected didSelect to be false (file is allowed, not disabled)")
	}
	if path != "" {
		t.Errorf("Expected empty path, got %s", path)
	}
	if len(updatedModel.Recents) != 2 {
		t.Errorf("Expected 2 recents in returned model, got %d", len(updatedModel.Recents))
	}
}

func TestModel_SelectedResetOnFileListChange(t *testing.T) {
	m := New()
	m.CurrentDirectory = t.TempDir()
	m.id = 1

	entries := make([]os.DirEntry, 50)
	for i := 0; i < 50; i++ {
		entries[i] = mockDirEntry{name: "file.txt"}
	}
	m.files = entries

	for i := 0; i < 48; i++ {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	}

	if m.selected != 48 {
		t.Fatalf("Expected selected to be 48 after 48 down presses, got %d", m.selected)
	}

	fewerEntries := make([]os.DirEntry, 5)
	for i := 0; i < 5; i++ {
		fewerEntries[i] = mockDirEntry{name: "file.txt"}
	}

	m, _ = m.Update(readDirMsg{id: 1, entries: fewerEntries})

	if m.selected != 0 {
		t.Errorf("Expected selected to be reset to 0 after readDirMsg with fewer entries, got %d", m.selected)
	}
	if m.min != 0 {
		t.Errorf("Expected min to be reset to 0, got %d", m.min)
	}
}
