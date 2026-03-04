package ui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// LoadingConfig holds configuration for rendering the loading view
type LoadingConfig struct {
	StartTime time.Time
	Theme     ThemePalette
}

// RenderLoading displays an arcade-style loading screen
func RenderLoading(cfg LoadingConfig) string {
	// Animated spinner frames
	frames := []string{"◜", "◝", "◞", "◟"}
	frameIndex := int(time.Since(cfg.StartTime).Seconds()*4) % 4
	frame := frames[frameIndex]

	// Use theme colors for loading
	theme := cfg.Theme
	spinnerColor := theme.TitleColor
	arcadeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.ArcadeGold)
	spinnerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(spinnerColor)
	panelStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(theme.TitleColor).
		Background(blendHex(string(theme.AppBg), string(theme.GlowColor), 0.22)).
		Padding(1, 3).
		Align(lipgloss.Center)
	subtitleStyle := lipgloss.NewStyle().
		Faint(true).
		Foreground(theme.AppFg)

	var s string
	s = "\n\n"
	content := spinnerStyle.Render(frame+" Initializing...") + "\n\n" +
		arcadeStyle.Render("INSERT COIN TO START") + "\n" +
		subtitleStyle.Render("(Loading tickets...)")
	s += panelStyle.Render(content)
	return s
}
