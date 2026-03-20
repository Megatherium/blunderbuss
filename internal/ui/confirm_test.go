package ui

import (
	"testing"

	"github.com/megatherium/blunderbust/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestConfirmView_ShowsReadyPanelAndLaunchButton(t *testing.T) {
	selection := domain.Selection{
		Ticket: domain.Ticket{
			ID:    "bb-123",
			Title: "Ticket Title",
		},
		Harness: domain.Harness{Name: "codex"},
		Model:   "model-a",
		Agent:   "build",
	}

	s := confirmView(selection, nil, false, "/tmp/worktree", TokyoNightTheme)

	assert.Contains(t, s, "Confirm Launch Spec")
	assert.Contains(t, s, "READY?")
	assert.Contains(t, s, "LAUNCH")
	assert.Contains(t, s, "[Press Enter to launch, e to edit, esc to go back]")
}

func TestConfirmView_ShowsDryRunBadge(t *testing.T) {
	selection := domain.Selection{
		Ticket:  domain.Ticket{ID: "bb-123", Title: "Ticket Title"},
		Harness: domain.Harness{Name: "codex"},
	}

	s := confirmView(selection, nil, true, "", TokyoNightTheme)
	assert.Contains(t, s, "[DRY RUN]")
}
