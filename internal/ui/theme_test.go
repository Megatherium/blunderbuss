package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestGetCyclingColor_UsesTokyoNightCycles(t *testing.T) {
	color := getCyclingColor(0.5, 0, &TokyoNightTheme)
	assert.Equal(t, "#7aa2f7", string(color))
}

func TestGetCyclingColor_UsesMatrixCyclesByDefault(t *testing.T) {
	color := getCyclingColor(0.5, 0, &MatrixTheme)
	assert.Equal(t, "#90EE90", string(color))
}

func TestGetCyclingColor_UsesCyberpunkCycles(t *testing.T) {
	color := getCyclingColor(0.5, 0, &CyberpunkTheme)
	assert.Equal(t, "#ff00ff", string(color))
}

func TestGetCyclingColor_NilThemeFallsBackToMatrixCycles(t *testing.T) {
	color := getCyclingColor(0.5, 0, nil)
	assert.Equal(t, "#90EE90", string(color))
}

func TestGetCyclingColor_InterpolatesBetweenDarkAndBase(t *testing.T) {
	color := getCyclingColor(0.25, 0, &TokyoNightTheme)
	assert.Equal(t, "#5b7dcc", string(color))
}

func TestGetCyclingColor_WrapsCycleIndex(t *testing.T) {
	// TokyoNight has 4 cycles; index 5 should wrap to cycle 1.
	color := getCyclingColor(0.5, 5, &TokyoNightTheme)
	assert.Equal(t, "#bb9af7", string(color))
}

func TestGetGlowColor_NonHexAppBackgroundFallsBackToGlowColor(t *testing.T) {
	theme := ThemePalette{
		Name:      "NonHexBg",
		AppBg:     lipgloss.Color("236"),
		GlowColor: lipgloss.Color("#445566"),
	}

	color := getGlowColor(0.7, &theme)
	assert.Equal(t, "#445566", string(color))
}
