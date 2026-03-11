package ui

import (
	"github.com/charmbracelet/bubbles/list"
)

func createErrorList(message string, theme ...*ThemePalette) list.Model {
	items := []list.Item{errorItem{message: message}}
	l := list.New(items, newGradientDelegate(theme...), 0, 0)
	l.Title = "Select a Ticket"
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	return l
}

type errorItem struct {
	message string
}

func (i errorItem) Title() string       { return "⚠ " + i.message }
func (i errorItem) Description() string { return "" }
func (i errorItem) FilterValue() string { return "" }

func isFocusedListFiltering(m UIModel) bool {
	if m.state != ViewStateMatrix {
		return false
	}

	switch m.focus {
	case FocusTickets:
		return m.ticketList.FilterState() == list.Filtering
	case FocusHarness:
		return m.harnessList.FilterState() == list.Filtering
	case FocusModel:
		return m.modelList.FilterState() == list.Filtering
	case FocusAgent:
		return m.agentList.FilterState() == list.Filtering
	}
	return false
}
