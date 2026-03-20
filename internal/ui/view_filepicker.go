package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/megatherium/blunderbust/internal/ui/filepicker"
)

// FilePickerConfig holds configuration for rendering the file picker view
type FilePickerConfig struct {
	Filepicker filepicker.Model
	Theme      ThemePalette
	Purpose    filePickerPurpose
}

// RenderFilePicker renders the file picker for adding projects or picking templates
func RenderFilePicker(cfg FilePickerConfig) string {
	var s strings.Builder

	theme := cfg.Theme

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TitleColor).
		MarginBottom(1)

	helpStyle := lipgloss.NewStyle().
		Faint(true).
		MarginTop(1)

	title := "Add Project - Select Directory"
	help := "Press 'a' to select highlighted directory, 'tab' to swap views, 'esc' to cancel"
	if cfg.Purpose == fpPurposeTemplate {
		title = "Pick Template File"
		help = "Press Enter to select file, 'ctrl+a' for all extensions, 'ctrl+.' for hidden files, 'esc' to cancel"
	}

	s.WriteString(titleStyle.Render(title))
	s.WriteString("\n\n")
	s.WriteString(cfg.Filepicker.View())
	s.WriteString("\n")
	s.WriteString(helpStyle.Render(help))

	return s.String()
}

// AddProjectConfig holds configuration for rendering the add project modal
type AddProjectConfig struct {
	PendingProjectPath string
	Theme              ThemePalette
}

// RenderAddProjectModal renders the confirmation modal for adding a project
func RenderAddProjectModal(cfg AddProjectConfig) string {
	var s strings.Builder

	theme := cfg.Theme

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TitleColor).
		MarginBottom(1)

	pathStyle := lipgloss.NewStyle().
		Foreground(theme.ReadyColor).
		Bold(true)

	helpStyle := lipgloss.NewStyle().
		Faint(true).
		MarginTop(1)

	s.WriteString(titleStyle.Render("Add Project?"))
	s.WriteString("\n\n")
	fmt.Fprintf(&s, "Add project at:\n%s", pathStyle.Render(cfg.PendingProjectPath))
	s.WriteString("\n\n")
	s.WriteString(helpStyle.Render("Press 'y' or Enter to confirm, 'n' or Esc to cancel"))

	return s.String()
}
