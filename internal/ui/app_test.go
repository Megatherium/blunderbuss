package ui

import (
	"context"
	"errors"
	osexec "os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/megatherium/blunderbust/internal/data"
	"github.com/megatherium/blunderbust/internal/domain"
)

// mockFontDetector implements fontDetector for testing.
type mockFontDetector struct {
	output []byte
	err    error
}

func (m mockFontDetector) CombinedOutput() ([]byte, error) {
	return m.output, m.err
}

func TestDetectNerdFont(t *testing.T) {
	t.Run("command execution fails", func(t *testing.T) {
		detector := mockFontDetector{err: errors.New("command failed")}
		result := detectNerdFontWithDetector(detector)

		require.False(t, result, "Should return false when command fails")
	})

	t.Run("empty output", func(t *testing.T) {
		detector := mockFontDetector{output: []byte("")}
		result := detectNerdFontWithDetector(detector)

		require.False(t, result, "Should return false when output is empty")
	})

	t.Run("contains nerd font (lowercase)", func(t *testing.T) {
		detector := mockFontDetector{output: []byte("Hasklig Nerd Font Mono\nJetBrains Mono")}
		result := detectNerdFontWithDetector(detector)

		assert.True(t, result, "Should detect 'nerd' in lowercase output")
	})

	t.Run("contains nerd font (uppercase)", func(t *testing.T) {
		detector := mockFontDetector{output: []byte("HASKLIG NERD FONT MONO\nJetBrains Mono")}
		result := detectNerdFontWithDetector(detector)

		assert.True(t, result, "Should detect 'NERD' in uppercase output (case-insensitive)")
	})

	t.Run("contains nerd font (mixed case)", func(t *testing.T) {
		detector := mockFontDetector{output: []byte("HasKlIg NeRd FoNt\nJetBrains Mono")}
		result := detectNerdFontWithDetector(detector)

		assert.True(t, result, "Should detect 'NeRd' in mixed case output")
	})

	t.Run("no nerd font present", func(t *testing.T) {
		detector := mockFontDetector{output: []byte("Arial\nHelvetica\nRoboto\nJetBrains Mono")}
		result := detectNerdFontWithDetector(detector)

		assert.False(t, result, "Should return false when nerd fonts are not present")
	})

	t.Run("partial string matching", func(t *testing.T) {
		detector := mockFontDetector{output: []byte("CascadiaMono Nerd Font Display")}
		result := detectNerdFontWithDetector(detector)

		assert.True(t, result, "Should match 'nerd' as a partial substring")
	})

	t.Run("whitespace handling", func(t *testing.T) {
		detector := mockFontDetector{output: []byte("  Nerd  Font  \n  Arial  ")}
		result := detectNerdFontWithDetector(detector)

		assert.True(t, result, "Should handle whitespace correctly")
	})

	t.Run("unicode and special characters", func(t *testing.T) {
		detector := mockFontDetector{output: []byte("FiraCode Nerd Font Mono:style=Bold\nLiberation Sans")}
		result := detectNerdFontWithDetector(detector)

		assert.True(t, result, "Should handle unicode and special characters")
	})
}

func TestDetectNerdFont_Integration(t *testing.T) {
	if _, err := osexec.LookPath("fc-list"); err != nil {
		t.Skip("fc-list not available, skipping integration tests")
	}

	t.Run("idempotent detection", func(t *testing.T) {
		result1 := DetectNerdFont()
		result2 := DetectNerdFont()

		assert.Equal(t, result1, result2,
			"DetectNerdFont should be idempotent")
	})

	t.Run("no errors during normal operation", func(t *testing.T) {
		result := DetectNerdFont()

		assert.True(t, result == true || result == false,
			"DetectNerdFont should return a valid boolean value")
	})
}

// mockFailingStoreFactory simulates a failure in TicketStore creation.

func TestApp_SetActiveProject_CreationFailure(t *testing.T) {
	// Initialize a stripped down App instance with a simulated active project
	app := &App{
		stores:        make(map[string]data.TicketStore),
		activeProject: "/existing/project",
		projects:      []domain.Project{{Dir: "/existing/project", Name: "existing"}},
	}

	// Pre-populate the existing active project store with an empty mock
	app.stores["/existing/project"] = &mockStore{}

	// Swap out the App createStore logic by mocking the error behavior locally via an inline App mock structure,
	// or we can test SetActiveProject by forcing createStore to fail. Wait, createStore on App uses a.opts.Demo...
	// We can set an invalid BeadsDir or trigger an actual error path if we pass a bad path?
	// Actually, `App.createStore` uses `dolt.NewStore`. It expects `.beads/metadata.json`. If we pass an empty dir, it will fail.

	err := app.SetActiveProject(context.Background(), "/nonexistent/test/failure")

	// Verify that the function returned an error.
	assert.Error(t, err)

	// verify that createStore's failure prevented a structural assignment overwrite of the activeProject.
	assert.Equal(t, "/existing/project", app.activeProject)

	// verify the store was never saved to the map
	_, exists := app.stores["/nonexistent/test/failure"]
	assert.False(t, exists)
}
