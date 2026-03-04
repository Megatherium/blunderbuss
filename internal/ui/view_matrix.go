package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// MatrixConfig holds all configuration needed to render the matrix view
type MatrixConfig struct {
	Width  int
	Height int

	ShowSidebar bool

	// Column widths
	SidebarWidth int
	TWidth       int
	HWidth       int
	MWidth       int
	AWidth       int

	// Column disabled states
	ModelColumnDisabled bool
	AgentColumnDisabled bool

	// Focus state
	Focus FocusColumn

	// Animation state
	AnimState AnimationState

	// Current theme
	Theme ThemePalette

	// List views
	TicketView  string
	HarnessView string
	ModelView   string
	AgentView   string
	SidebarView string

	// List titles for focus indicators
	TicketTitle  string
	HarnessTitle string
	ModelTitle   string
	AgentTitle   string
}

// RenderMatrix renders the main matrix view with 4 columns
func RenderMatrix(cfg MatrixConfig) string {
	// Guard against uninitialized dimensions
	if cfg.Height < filterHeight+2 {
		return "Initializing..."
	}

	listHeight := cfg.Height - filterHeight

	theme := cfg.Theme

	activeColor := getActiveColor(cfg.AnimState, cfg.Focus, theme)
	glowColor := getGlowColor(cfg.AnimState.PulsePhase, &theme)

	activeBorder := createActiveBorder(listHeight, activeColor, glowColor, theme)
	inactiveBorder := createInactiveBorder(listHeight, theme)

	capView := func(view string, w int) string {
		return lipgloss.NewStyle().MaxHeight(listHeight - 2).MaxWidth(w - 2).Render(view)
	}
	faintCapView := func(view string, w int) string {
		return lipgloss.NewStyle().Faint(true).MaxHeight(listHeight - 2).MaxWidth(w - 2).Render(view)
	}

	// Render columns
	tView := renderMatrixColumn(cfg.TicketView, cfg.TWidth, cfg.Focus == FocusTickets,
		cfg.TicketTitle, theme, activeBorder, inactiveBorder, capView, faintCapView)

	hView := renderMatrixColumn(cfg.HarnessView, cfg.HWidth, cfg.Focus == FocusHarness,
		cfg.HarnessTitle, theme, activeBorder, inactiveBorder, capView, faintCapView)

	mView := renderModelColumn(cfg, theme, listHeight, capView, faintCapView)
	aView := renderAgentColumn(cfg, theme, listHeight, capView, faintCapView)

	matrixWidth := cfg.TWidth + cfg.HWidth + cfg.MWidth + cfg.AWidth + 6

	filterLabel := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.FocusIndicator).
		Render("Filters:")
	filterHint := lipgloss.NewStyle().
		Faint(true).
		Foreground(theme.AppFg).
		Render("(Press / to search)")
	filterContent := lipgloss.JoinHorizontal(lipgloss.Left, filterLabel, " [All]  |  ", filterHint)

	filterBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(theme.TitleColor).
		Background(blendHex(string(theme.AppBg), string(theme.GlowColor), 0.25)).
		Width(matrixWidth-2).
		Height(1).
		Padding(0, 1).
		Render(filterContent)

	matrixBox := lipgloss.JoinHorizontal(lipgloss.Top,
		tView,
		lipgloss.NewStyle().Width(2).Render("  "),
		hView,
		lipgloss.NewStyle().Width(2).Render("  "),
		mView,
		lipgloss.NewStyle().Width(2).Render("  "),
		aView,
	)

	rightPanelBox := lipgloss.JoinVertical(lipgloss.Top, filterBox, matrixBox)

	if cfg.ShowSidebar {
		return applySidebarBorder(cfg, rightPanelBox, activeColor)
	}

	return rightPanelBox
}

func getActiveColor(animState AnimationState, focus FocusColumn, theme ThemePalette) lipgloss.Color {
	if animState.shouldShowFlash(focus) {
		return FlashColor
	}
	return getCyclingColor(animState.PulsePhase, animState.ColorCycleIndex, &theme)
}

func createActiveBorder(listHeight int, activeColor, glowColor lipgloss.Color, theme ThemePalette) func(int) lipgloss.Style {
	return func(w int) lipgloss.Style {
		if w < 2 {
			w = 2
		}
		return lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(activeColor).
			Background(blendHex(string(theme.AppBg), string(glowColor), 0.45)).
			Foreground(theme.AppFg).
			Padding(0, 1).
			Width(w - 2).
			Height(listHeight - 2)
	}
}

func createInactiveBorder(listHeight int, theme ThemePalette) func(int) lipgloss.Style {
	return func(w int) lipgloss.Style {
		if w < 2 {
			w = 2
		}
		return lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(ThemeInactive).
			Background(blendHex(string(theme.AppBg), string(theme.GlowColor), 0.12)).
			Foreground(theme.AppFg).
			Faint(false).
			Padding(0, 1).
			Width(w - 2).
			Height(listHeight - 2)
	}
}

func renderMatrixColumn(
	view string, width int,
	isFocused bool,
	title string,
	theme ThemePalette,
	activeBorder, inactiveBorder func(int) lipgloss.Style,
	capView, faintCapView func(string, int) string,
) string {
	const focusIndicator = "▶ "
	const noIndicator = "  "

	focusedTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.FocusIndicator)
	focusedTitleBadgeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.AppBg).
		Background(theme.TitleColor).
		Padding(0, 1)
	inactiveTitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TitleColor).
		Faint(true)

	if isFocused {
		indicator := focusedTitleStyle.Render(focusIndicator)
		titledView := indicator + focusedTitleBadgeStyle.Render(title) + "\n" + view
		return activeBorder(width).Render(capView(titledView, width))
	}

	titledView := noIndicator + inactiveTitleStyle.Render(title) + "\n" + view
	return inactiveBorder(width).Render(faintCapView(titledView, width))
}

func renderModelColumn(cfg MatrixConfig, theme ThemePalette, listHeight int,
	capView, faintCapView func(string, int) string) string {
	if cfg.ModelColumnDisabled {
		disabledStyle := lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(ThemeInactive).
			Background(blendHex(string(theme.AppBg), string(theme.GlowColor), 0.08)).
			Faint(true).
			Width(cfg.MWidth-2).
			Height(listHeight-2).
			Padding(0, 1).
			Align(lipgloss.Center, lipgloss.Center)
		return disabledStyle.Render("Models\n\nN/A\n\nNo models available\nfor this harness")
	}

	return renderMatrixColumn(cfg.ModelView, cfg.MWidth, cfg.Focus == FocusModel,
		"Models", theme,
		func(w int) lipgloss.Style {
			return createActiveBorder(listHeight, getActiveColor(cfg.AnimState, FocusModel, theme), getGlowColor(cfg.AnimState.PulsePhase, &theme), theme)(w)
		},
		createInactiveBorder(listHeight, theme),
		capView, faintCapView)
}

func renderAgentColumn(cfg MatrixConfig, theme ThemePalette, listHeight int,
	capView, faintCapView func(string, int) string) string {
	if cfg.AgentColumnDisabled {
		disabledStyle := lipgloss.NewStyle().
			Border(lipgloss.ThickBorder()).
			BorderForeground(ThemeInactive).
			Background(blendHex(string(theme.AppBg), string(theme.GlowColor), 0.08)).
			Faint(true).
			Width(cfg.AWidth-2).
			Height(listHeight-2).
			Padding(0, 1).
			Align(lipgloss.Center, lipgloss.Center)
		return disabledStyle.Render("Agents\n\nN/A\n\nNo agents available\nfor this harness")
	}

	return renderMatrixColumn(cfg.AgentView, cfg.AWidth, cfg.Focus == FocusAgent,
		"Agents", theme,
		func(w int) lipgloss.Style {
			return createActiveBorder(listHeight, getActiveColor(cfg.AnimState, FocusAgent, theme), getGlowColor(cfg.AnimState.PulsePhase, &theme), theme)(w)
		},
		createInactiveBorder(listHeight, theme),
		capView, faintCapView)
}

func applySidebarBorder(cfg MatrixConfig, rightPanelBox string, activeColor lipgloss.Color) string {
	w := cfg.SidebarWidth
	if w < 2 {
		w = 2
	}

	sidebarBorder := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		Background(blendHex(string(cfg.Theme.AppBg), string(cfg.Theme.GlowColor), 0.16)).
		Foreground(cfg.Theme.AppFg).
		Padding(0, 1).
		Width(w - 2).
		Height(cfg.Height - 2)

	if cfg.Focus == FocusSidebar {
		sidebarBorder = sidebarBorder.BorderForeground(activeColor)
	} else {
		sidebarBorder = sidebarBorder.BorderForeground(ThemeInactive)
	}

	sidebarBox := sidebarBorder.Render(cfg.SidebarView)
	return lipgloss.JoinHorizontal(lipgloss.Top, sidebarBox,
		lipgloss.NewStyle().Width(2).Render("  "), rightPanelBox)
}

func blendHex(baseHex, accentHex string, ratio float64) lipgloss.Color {
	if ratio <= 0 {
		return lipgloss.Color(baseHex)
	}
	if ratio > 1 {
		ratio = 1
	}

	br, bg, bb, baseErr := parseHexColor(baseHex)
	ar, ag, ab, accentErr := parseHexColor(accentHex)
	if baseErr != nil || accentErr != nil {
		return lipgloss.Color(baseHex)
	}

	mix := func(base, accent uint8, t float64) uint8 {
		return uint8(float64(base) + (float64(accent)-float64(base))*t)
	}

	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x",
		mix(br, ar, ratio),
		mix(bg, ag, ratio),
		mix(bb, ab, ratio),
	))
}
