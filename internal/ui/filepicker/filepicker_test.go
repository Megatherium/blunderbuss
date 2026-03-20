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
