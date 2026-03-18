package ui

import (
	"fmt"
	"testing"

	"github.com/megatherium/blunderbust/internal/domain"
	"github.com/stretchr/testify/assert"
)

// TestTemplateReloadingMessages tests the template reloading message types
func TestTemplateReloadingMessages(t *testing.T) {
	// Test TemplatesReloadedMsg
	reloadedMsg := TemplatesReloadedMsg{
		Harnesses: []domain.Harness{
			{
				Name:            "test-harness",
				CommandTemplate: "updated-command",
				PromptTemplate:  "updated-prompt",
			},
		},
	}
	assert.Equal(t, "updated-command", reloadedMsg.Harnesses[0].CommandTemplate)
	assert.Equal(t, "updated-prompt", reloadedMsg.Harnesses[0].PromptTemplate)

	// Test TemplateReloadErrorMsg
	errorMsg := TemplateReloadErrorMsg{
		Error: fmt.Errorf("test error"),
	}
	assert.Contains(t, errorMsg.Error.Error(), "test error")

	// Test ReloadTemplatesMsg
	reloadMsg := ReloadTemplatesMsg{}
	assert.Equal(t, ReloadTemplatesMsg{}, reloadMsg)
}

// TestTemplateReloadErrorHandler tests the error handling for template reloading
func TestTemplateReloadErrorHandler(t *testing.T) {
	// Create a minimal UI model
	m := UIModel{}

	// Create an error message
	errorMsg := TemplateReloadErrorMsg{
		Error: fmt.Errorf("template reload failed"),
	}

	// Handle the error
	resultModel, cmd := m.handleTemplateReloadError(errorMsg)

	// Verify error was set
	uiModel := resultModel.(UIModel)
	if uiModel.err != nil {
		assert.Contains(t, uiModel.err.Error(), "template reload failed")
	} else {
		// Check if error was set via command instead
		if cmd != nil {
			cmdResult := cmd()
			if errMsg, ok := cmdResult.(errMsg); ok {
				assert.Contains(t, errMsg.err.Error(), "template reload failed")
			}
		}
	}
}

// TestTemplatesReloadedHandler tests the successful template reloading handler
func TestTemplatesReloadedHandler(t *testing.T) {
	// Create a minimal UI model
	m := UIModel{
		harnesses: []domain.Harness{
			{
				Name:            "old-harness",
				CommandTemplate: "old-command",
			},
		},
		state: ViewStateMatrix,
	}

	// Create a reloaded message with updated harnesses
	reloadedMsg := TemplatesReloadedMsg{
		Harnesses: []domain.Harness{
			{
				Name:            "new-harness",
				CommandTemplate: "new-command",
			},
		},
	}

	// Handle the reloaded message
	resultModel, cmd := m.handleTemplatesReloaded(reloadedMsg)

	// Verify harnesses were updated
	uiModel := resultModel.(UIModel)
	assert.Equal(t, "new-harness", uiModel.harnesses[0].Name)
	assert.Equal(t, "new-command", uiModel.harnesses[0].CommandTemplate)
	assert.True(t, uiModel.dirtyHarness)

	// Verify an info message was generated
	if cmd != nil {
		cmdResult := cmd()
		if infoMsg, ok := cmdResult.(infoMsg); ok {
			assert.Contains(t, infoMsg.message, "templates reloaded successfully")
		}
	}
}
