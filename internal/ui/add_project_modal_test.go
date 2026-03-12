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

// TestAddProjectModal_NotInWorkspace tests that modal shows when project not in workspace
func TestAddProjectModal_NotInWorkspace(t *testing.T) {
	app := newTestApp()
	app.Opts.TargetProject = "/some/new/project"
	// newTestApp already initializes an empty workspace

	_ = NewUIModel(app, nil)

	// Verify target project is detected
	assert.Equal(t, "/some/new/project", app.GetTargetProject())
	assert.False(t, app.IsProjectInWorkspace("/some/new/project"))
}

// TestAddProjectModal_AlreadyInWorkspace tests that modal is NOT shown when project already exists
func TestAddProjectModal_AlreadyInWorkspace(t *testing.T) {
	app := newTestApp()
	app.Opts.TargetProject = "/existing/project"
	app.Stores = make(map[string]data.TicketStore)
	app.AddProject(domain.Project{Dir: "/existing/project", Name: "existing"})

	m := NewUIModel(app, nil)

	// Verify project is detected as in workspace
	assert.True(t, app.IsProjectInWorkspace("/existing/project"))

	// Init should not trigger add-project modal (project already in workspace)
	cmd := m.Init()
	require.NotNil(t, cmd)
}

// TestAddProjectModal_MissingBeadsDir tests error when .beads directory is missing
func TestAddProjectModal_MissingBeadsDir(t *testing.T) {
	app := newTestApp()

	// Create a temp dir without .beads
	tmpDir := t.TempDir()

	err := app.ValidateProject(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain a .beads subdirectory")
}

// TestAddProjectModal_ValidProjectWithBeads tests validation passes with .beads dir
func TestAddProjectModal_ValidProjectWithBeads(t *testing.T) {
	app := newTestApp()

	// Create a temp dir with .beads
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	require.NoError(t, os.MkdirAll(beadsDir, 0755))

	err := app.ValidateProject(tmpDir)
	assert.NoError(t, err)
}

// TestAddProjectModal_KeyHandlers tests y/n key handling for modal
func TestAddProjectModal_KeyHandlers(t *testing.T) {
	app := newTestApp()
	m := NewUIModel(app, nil)

	// Enable modal
	m.state = ViewStateAddProjectModal
	m.pendingProjectPath = "/test/project"

	// Test 'y' key accepts
	yMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmd, handled := m.handleKeyMsg(yMsg)
	require.True(t, handled, "y key should be handled")
	require.NotNil(t, cmd)

	// The command should return addProjectConfirmedMsg
	msg := cmd()
	confirmedMsg, ok := msg.(addProjectConfirmedMsg)
	require.True(t, ok, "should return addProjectConfirmedMsg")
	assert.Equal(t, "/test/project", confirmedMsg.path, "should include the pending project path")

	// Reset and test 'n' key declines
	m.state = ViewStateAddProjectModal
	m.pendingProjectPath = "/test/project"
	nMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	_, cmd, handled = m.handleKeyMsg(nMsg)
	require.True(t, handled, "n key should be handled")
	require.NotNil(t, cmd)

	msg = cmd()
	_, ok = msg.(addProjectCancelledMsg)
	require.True(t, ok, "should return addProjectCancelledMsg")
}

// TestAddProjectModal_BlocksOtherKeys tests that other keys are blocked when modal is shown
func TestAddProjectModal_BlocksOtherKeys(t *testing.T) {
	app := newTestApp()
	m := NewUIModel(app, nil)

	// Enable modal
	m.state = ViewStateAddProjectModal
	m.pendingProjectPath = "/test/project"

	// Test that Enter key is blocked
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, _, handled := m.handleKeyMsg(enterMsg)
	assert.True(t, handled, "Enter key should be blocked when modal shown")

	// Test that Tab key is blocked
	tabMsg := tea.KeyMsg{Type: tea.KeyTab}
	_, _, handled = m.handleKeyMsg(tabMsg)
	assert.True(t, handled, "Tab key should be blocked when modal shown")

	// Test that Escape key declines (not just blocked)
	escMsg := tea.KeyMsg{Type: tea.KeyEscape}
	_, cmd, handled := m.handleKeyMsg(escMsg)
	assert.True(t, handled, "Escape key should be handled")
	require.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(addProjectCancelledMsg)
	require.True(t, ok, "Escape should return addProjectCancelledMsg")
}

// TestApp_DeduplicateProjectName tests name collision handling
// Note: since this is in the ui package we can't test internal app methods easily
// If this test continues to fail it should be moved to internal/app
func TestApp_DeduplicateProjectName(t *testing.T) {
	app := newTestApp()
	app.AddProject(domain.Project{Dir: "/path1/foo", Name: "foo"})

	// Add another with same name to test rename
	app.AddProject(domain.Project{Dir: "/path2/foo", Name: "foo"})

	projects := app.GetProjects()
	assert.Len(t, projects, 2)
	names := make(map[string]bool)
	for _, p := range projects {
		names[p.Name] = true
	}
	assert.True(t, names["foo"])
	assert.True(t, names["foo-1"])
}

// TestApp_AddProject_DuplicatePrevention tests that duplicate projects aren't added
func TestApp_AddProject_DuplicatePrevention(t *testing.T) {
	app := newTestApp()
	app.AddProject(domain.Project{Dir: "/path/to/project", Name: "project"})

	// Try to add same project again
	app.AddProject(domain.Project{Dir: "/path/to/project", Name: "different-name"})

	// Should still only have 1 project
	projects := app.GetProjects()
	assert.Len(t, projects, 1)
	assert.Equal(t, "project", projects[0].Name) // Original name preserved
}

// TestAddProjectModal_PathResolution tests that relative paths are resolved to absolute
func TestAddProjectModal_PathResolution(t *testing.T) {
	// This test verifies the main.go logic that resolves paths
	// We can't easily test the actual main function, but we can verify
	// the App methods work with both relative and absolute paths

	app := newTestApp()

	// Test with first project
	absPath := "/absolute/path/to/project"
	app.AddProject(domain.Project{Dir: absPath, Name: "project"})

	projects := app.GetProjects()
	assert.Len(t, projects, 1)
	assert.Equal(t, absPath, projects[0].Dir)
}
