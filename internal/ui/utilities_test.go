package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateErrorList(t *testing.T) {
	l := createErrorList("Test error message")

	assert.Equal(t, "Select a Ticket", l.Title)
	assert.False(t, l.ShowTitle())
	assert.False(t, l.ShowStatusBar())

	items := l.Items()
	require.Len(t, items, 1)

	errItem, ok := items[0].(errorItem)
	require.True(t, ok, "Expected errorItem type")
	assert.Equal(t, "⚠ Test error message", errItem.Title())
	assert.Equal(t, "", errItem.Description())
	assert.Equal(t, "", errItem.FilterValue())
}

func TestErrorItem(t *testing.T) {
	err := errorItem{message: "Something went wrong"}

	assert.Equal(t, "⚠ Something went wrong", err.Title())
	assert.Equal(t, "", err.Description())
	assert.Equal(t, "", err.FilterValue())
}

func TestIsFocusedListFiltering(t *testing.T) {
	app := newTestApp()
	m := NewUIModel(app, nil)

	// Initially not in matrix state
	assert.False(t, isFocusedListFiltering(m))

	// Set to matrix state - not filtering
	m.state = ViewStateMatrix
	m.focus = FocusTickets
	assert.False(t, isFocusedListFiltering(m))

	// Focus on different columns
	m.focus = FocusHarness
	assert.False(t, isFocusedListFiltering(m))

	m.focus = FocusModel
	assert.False(t, isFocusedListFiltering(m))

	m.focus = FocusAgent
	assert.False(t, isFocusedListFiltering(m))

	// Not in matrix state
	m.state = ViewStateLoading
	assert.False(t, isFocusedListFiltering(m))
}
