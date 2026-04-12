package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLayoutDimensions_Compute_WithSidebar(t *testing.T) {
	termW, termH := 160, 30
	layout := Compute(termW, termH, true, false)

	assert.Equal(t, Grid4x1, layout.GridMode)
	assert.Equal(t, 160, layout.TermWidth)
	assert.Equal(t, 30, layout.TermHeight)
	assert.NotZero(t, layout.SidebarWidth)
	assert.Equal(t, layout.Width, layout.SidebarWidth+layout.TWidth+layout.HWidth+layout.MWidth+layout.AWidth+8)
}

func TestLayoutDimensions_Compute_WithoutSidebar(t *testing.T) {
	termW, termH := 160, 30
	layout := Compute(termW, termH, false, false)

	assert.Equal(t, Grid4x1, layout.GridMode)
	assert.Equal(t, 160, layout.TermWidth)
	assert.Equal(t, 30, layout.TermHeight)
	assert.Equal(t, 0, layout.SidebarWidth)
	assert.Equal(t, layout.Width, layout.TWidth+layout.HWidth+layout.MWidth+layout.AWidth+6)
}

func TestLayoutDimensions_Compute_MinimumDimensions(t *testing.T) {
	layout := Compute(10, 5, true, false)

	assert.Equal(t, minWindowWidth, layout.Width)
	assert.Equal(t, minWindowHeight, layout.Height)
}

func TestLayoutDimensions_Compute_MinAgentWidth(t *testing.T) {
	layout := Compute(160, 30, true, false)

	assert.GreaterOrEqual(t, layout.AWidth, minAgentWidth)
}

func TestLayoutDimensions_Compute_HarnessColumnIsHalf(t *testing.T) {
	layout := Compute(160, 30, true, false)

	assert.Equal(t, layout.HWidth, layout.TWidth/2)
}

func TestLayoutDimensions_Compute_InnerListHeight(t *testing.T) {
	layout := Compute(160, 30, true, false)

	expectedHeight := layout.Height - filterHeight - borderWidth - 1
	if expectedHeight < 1 {
		expectedHeight = 1
	}
	assert.Equal(t, expectedHeight, layout.InnerTHeight)
}

func TestLayoutDimensions_Compute_PureFunction(t *testing.T) {
	layout1 := Compute(160, 30, true, false)
	layout2 := Compute(160, 30, true, false)

	assert.Equal(t, layout1, layout2)
}

func TestLayoutDimensions_Compute_ShowSidebarChangesWidths(t *testing.T) {
	layoutWithSidebar := Compute(160, 30, true, false)
	layoutWithoutSidebar := Compute(160, 30, false, false)

	assert.NotEqual(t, layoutWithSidebar.SidebarWidth, layoutWithoutSidebar.SidebarWidth)
	assert.Greater(t, layoutWithSidebar.SidebarWidth, 0)
	assert.Equal(t, 0, layoutWithoutSidebar.SidebarWidth)
}

func TestLayoutDimensions_Compute_2x2(t *testing.T) {
	termW, termH := 80, 30 // between 60 and 110
	layout := Compute(termW, termH, false, false)

	assert.Equal(t, Grid2x2, layout.GridMode)
	assert.Equal(t, layout.TWidth, layout.MWidth)
	assert.Equal(t, layout.HWidth, layout.AWidth)
}

func TestLayoutDimensions_Compute_1x4(t *testing.T) {
	termW, termH := 50, 30 // < 60
	layout := Compute(termW, termH, false, false)

	assert.Equal(t, Grid1x4, layout.GridMode)
	assert.Equal(t, layout.TWidth, layout.HWidth)
	assert.Equal(t, layout.HWidth, layout.MWidth)
	assert.Equal(t, layout.MWidth, layout.AWidth)
}

func TestLayoutDimensions_Compute_BoundaryValues(t *testing.T) {
	layout64 := Compute(64, 30, false, false)
	assert.Equal(t, Grid2x2, layout64.GridMode)

	layout115 := Compute(115, 30, false, false)
	assert.Equal(t, Grid4x1, layout115.GridMode)
}

func TestLayoutDimensions_Compute_SidebarSuppression(t *testing.T) {
	// 2x2 mode with sidebar requested -> sidebar should be suppressed
	layout2x2 := Compute(80, 30, true, false)
	assert.Equal(t, Grid2x2, layout2x2.GridMode)
	assert.Equal(t, 0, layout2x2.SidebarWidth)

	// 1x4 mode with sidebar requested -> sidebar should be suppressed
	layout1x4 := Compute(50, 30, true, false)
	assert.Equal(t, Grid1x4, layout1x4.GridMode)
	assert.Equal(t, 0, layout1x4.SidebarWidth)
}

func TestLayoutDimensions_Compute_2x2Zoom(t *testing.T) {
	layout := Compute(80, 30, false, true)
	assert.Equal(t, Grid2x2, layout.GridMode)
	
	// Harness should be zoomed to minimum
	assert.Equal(t, minZoomedColumnWidth, layout.HWidth)
	
	// Ticket should take the rest of the top row
	expectedUsable := layout.Width - marginWithoutSidebar - 2 // -2 for gap
	assert.Equal(t, expectedUsable-minZoomedColumnWidth, layout.TWidth)

	// Bottom row should be evenly split
	expectedTopHalf := expectedUsable / 2
	assert.Equal(t, expectedTopHalf, layout.MWidth)
	assert.Equal(t, expectedUsable-expectedTopHalf, layout.AWidth)
}

func TestLayoutDimensions_Compute_Heights(t *testing.T) {
	// Check heights for 2x2
	layout2x2 := Compute(80, 30, false, false)
	listHeight2x2 := layout2x2.Height - filterHeight
	listHeight2x2 -= 1 // gap
	expectedRow1 := listHeight2x2 / 2
	expectedRow2 := listHeight2x2 - expectedRow1
	
	assert.Equal(t, expectedRow1, layout2x2.THeight)
	assert.Equal(t, expectedRow1, layout2x2.HHeight)
	assert.Equal(t, expectedRow2, layout2x2.MHeight)
	assert.Equal(t, expectedRow2, layout2x2.AHeight)

	// Check heights for 1x4
	layout1x4 := Compute(50, 30, false, false)
	listHeight1x4 := layout1x4.Height - filterHeight
	listHeight1x4 -= 3 // gaps
	expectedRow1_1x4 := listHeight1x4 / 4
	expectedRow2_1x4 := listHeight1x4 / 4
	expectedRow3_1x4 := listHeight1x4 / 4
	expectedRow4_1x4 := listHeight1x4 - (expectedRow1_1x4 * 3)

	assert.Equal(t, expectedRow1_1x4, layout1x4.THeight)
	assert.Equal(t, expectedRow2_1x4, layout1x4.HHeight)
	assert.Equal(t, expectedRow3_1x4, layout1x4.MHeight)
	assert.Equal(t, expectedRow4_1x4, layout1x4.AHeight)
}
