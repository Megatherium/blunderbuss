package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"

	"github.com/megatherium/blunderbust/internal/config"
	"github.com/megatherium/blunderbust/internal/data"
	"github.com/megatherium/blunderbust/internal/data/dolt"
	"github.com/megatherium/blunderbust/internal/domain"
	"github.com/megatherium/blunderbust/internal/ui/filepicker"
)

// MainContentConfig holds configuration for rendering the main content
type MainContentConfig struct {
	State              ViewState
	Focus              FocusColumn
	ViewingAgentID     string
	Selection          domain.Selection
	Renderer           *config.Renderer
	DryRun             bool
	SelectedWorktree   string
	CurrentTheme       ThemePalette
	ShowModal          bool
	ModalContent       string
	PendingProjectPath string
	Warnings           []string
	Width              int
	Height             int
	Err                error
	RetryStore         data.TicketStore

	// View dependencies
	MatrixConfig      MatrixConfig
	Agent             *RunningAgent
	Filepicker        filepicker.Model
	FilePickerPurpose filePickerPurpose
	AnimState         AnimationState

	// Inline edit state
	InlineEditTextarea textarea.Model
	InlineEditMode     editMode
	InlineEditError    string
}

// RenderMainContent renders the main content area based on current state
func RenderMainContent(cfg MainContentConfig) string {
	var s string

	switch cfg.State {
	case ViewStateLoading:
		s = RenderLoading(LoadingConfig{
			StartTime: cfg.AnimState.StartTime,
			Theme:     cfg.CurrentTheme,
		})
	case ViewStateFilePicker:
		s = RenderFilePicker(FilePickerConfig{
			Filepicker: cfg.Filepicker,
			Theme:      cfg.CurrentTheme,
			Purpose:    cfg.FilePickerPurpose,
		})
	case ViewStateAddProjectModal:
		s = RenderAddProjectModal(AddProjectConfig{
			PendingProjectPath: cfg.PendingProjectPath,
			Theme:              cfg.CurrentTheme,
		})
	case ViewStateAgentOutput:
		s = RenderAgentOutput(AgentConfig{
			Agent:  cfg.Agent,
			Width:  cfg.Width,
			Height: cfg.Height,
			Theme:  cfg.CurrentTheme,
		})
	case ViewStateMatrix:
		s = RenderMatrix(cfg.MatrixConfig)
	case ViewStateConfirm:
		s = confirmView(cfg.Selection, cfg.Renderer, cfg.DryRun, cfg.SelectedWorktree, cfg.CurrentTheme)
	case ViewStateInlineEdit:
		s = RenderInlineEdit(InlineEditConfig{
			Textarea: cfg.InlineEditTextarea,
			Mode:     cfg.InlineEditMode,
			Theme:    cfg.CurrentTheme,
			Error:    cfg.InlineEditError,
			Width:    cfg.Width,
			Height:   cfg.Height,
		})
	case ViewStateError:
		s = renderErrorState(cfg)
	}

	// Overlay modals on top
	s = renderModalOverlay(s, cfg)
	s = renderWarnings(s, cfg.Warnings)

	return s
}

func renderErrorState(cfg MainContentConfig) string {
	hasRetry := false
	hasStart := false
	if cfg.RetryStore != nil {
		hasRetry = true
		if doltStore, ok := cfg.RetryStore.(*dolt.Store); ok {
			hasStart = doltStore.CanRetryConnection()
		}
	}
	return errorView(cfg.Err, hasRetry, hasStart)
}

func renderModalOverlay(content string, cfg MainContentConfig) string {
	if !cfg.ShowModal {
		return content
	}

	modalWidth := cfg.Width - 10
	if modalWidth < 40 {
		modalWidth = 40
	}

	modalBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ThemeActive).
		Padding(1, 2).
		Width(modalWidth).
		Render(cfg.ModalContent)

	return lipgloss.Place(cfg.Width, cfg.Height, lipgloss.Center, lipgloss.Center, modalBox)
}

func renderWarnings(content string, warnings []string) string {
	if len(warnings) == 0 {
		return content
	}

	warningStyle := lipgloss.NewStyle().Foreground(ThemeWarning).MarginTop(1)
	for _, w := range warnings {
		content += "\n" + warningStyle.Render("⚠ "+w)
	}
	return content
}
