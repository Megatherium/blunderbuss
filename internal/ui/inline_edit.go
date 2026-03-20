package ui

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"
)

type InlineEditConfig struct {
	Textarea textarea.Model
	Mode     editMode
	Theme    ThemePalette
	Error    string
	Width    int
	Height   int
}

func RenderInlineEdit(cfg InlineEditConfig) string {
	themeTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(cfg.Theme.TitleColor).
		MarginBottom(1)

	titleText := "Edit Command Template"
	if cfg.Mode == editModePrompt {
		titleText = "Edit Prompt Template"
	}

	s := themeTitleStyle.Render(titleText) + "\n\n"
	s += itemStyle.Render("Enter = newline, Ctrl-y = accept, Esc = cancel") + "\n\n"

	if cfg.Error != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)
		s += errorStyle.Render("Error: "+cfg.Error) + "\n\n"
	}

	taStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(cfg.Theme.ReadyColor).
		Padding(1, 2)

	s += taStyle.Render(cfg.Textarea.View()) + "\n"

	return s
}
