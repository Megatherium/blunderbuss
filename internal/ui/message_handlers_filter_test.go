package ui

import (
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/stretchr/testify/assert"

	"github.com/megatherium/blunderbust/internal/data"
	"github.com/megatherium/blunderbust/internal/domain"
)

func TestHandleTicketsLoaded_PreservesSelectionWithFilter(t *testing.T) {
	app := newTestApp()
	app.ActiveProject = "."
	app.Stores = make(map[string]data.TicketStore)
	app.Stores["."] = &mockStore{}

	m := NewUIModel(app, nil)

	// Set up tickets and filter (simulating user has filtered list)
	tickets := []domain.Ticket{
		{ID: "bb-1", Title: "Apple Task", Status: "open", Priority: 1},
		{ID: "bb-2", Title: "Banana Task", Status: "open", Priority: 2},
		{ID: "bb-3", Title: "Cherry Task", Status: "open", Priority: 3},
	}
	m.ticketList = newTicketList(tickets)

	// Manually set filter state to simulate user having filtered the list
	// (This is what happens after a previous handleTicketsLoaded call)
	// Filter "n" matches only Banana (contains "n" in Title)
	m.ticketList.SetFilterText("n")
	m.ticketList.SetFilterState(list.FilterApplied)

	// Select "Banana Task" (should be at index 0 in filtered results: Banana)
	m.ticketList.Select(0)
	selectedItem, _ := m.ticketList.SelectedItem().(ticketItem)
	m.selection.Ticket = selectedItem.ticket

	// Refresh with same tickets (simulating refresh with active filter)
	msg := ticketsLoadedMsg(tickets)
	updatedModel, _ := m.handleTicketsLoaded(msg)
	updatedM := updatedModel.(UIModel)

	// Selection should be preserved in filtered list
	assert.Equal(t, "bb-2", updatedM.selection.Ticket.ID)
	assert.Equal(t, "Banana Task", updatedM.selection.Ticket.Title)

	// List cursor should be at same filtered index
	newSelectedItem, ok := updatedM.ticketList.SelectedItem().(ticketItem)
	assert.True(t, ok)
	assert.Equal(t, "bb-2", newSelectedItem.ticket.ID)

	// Filter should still be active
	assert.Equal(t, list.FilterApplied, updatedM.ticketList.FilterState())
	assert.Equal(t, "n", updatedM.ticketList.FilterValue())

	// Verify filtered items are correct
	visibleItems := updatedM.ticketList.VisibleItems()
	assert.Equal(t, 1, len(visibleItems))
}

func TestHandleTicketsLoaded_ClearsSelectionWhenExcludedByFilter(t *testing.T) {
	app := newTestApp()
	app.ActiveProject = "."
	app.Stores = make(map[string]data.TicketStore)
	app.Stores["."] = &mockStore{}

	m := NewUIModel(app, nil)

	// Set up tickets
	tickets := []domain.Ticket{
		{ID: "bb-1", Title: "Apple Ticket", Status: "open", Priority: 1},
		{ID: "bb-2", Title: "Banana Ticket", Status: "open", Priority: 2},
		{ID: "bb-3", Title: "Cherry Ticket", Status: "open", Priority: 3},
	}
	m.ticketList = newTicketList(tickets)

	// Apply filter that matches "Apple"
	m.ticketList.SetFilterText("Apple")
	m.ticketList.SetFilterState(list.FilterApplied)

	// Select "Apple Ticket" (only result)
	m.ticketList.Select(0)
	selectedItem, _ := m.ticketList.SelectedItem().(ticketItem)
	m.selection.Ticket = selectedItem.ticket

	// Refresh tickets where "Apple" ticket is excluded by filter
	// Change filter to "Zebra" which matches nothing
	tickets[0].Title = "Apricot Ticket" // Change title so Apple is excluded
	msg := ticketsLoadedMsg(tickets)
	updatedModel, _ := m.handleTicketsLoaded(msg)
	updatedM := updatedModel.(UIModel)

	// Previous selection should be cleared
	assert.Equal(t, "", updatedM.selection.Ticket.ID)

	// Filter should be preserved
	assert.Equal(t, list.FilterApplied, updatedM.ticketList.FilterState())
	assert.Equal(t, "Apple", updatedM.ticketList.FilterValue())
}

func TestHandleTicketsLoaded_PreservesSelectionWithPositionChange(t *testing.T) {
	app := newTestApp()
	app.ActiveProject = "."
	app.Stores = make(map[string]data.TicketStore)
	app.Stores["."] = &mockStore{}

	m := NewUIModel(app, nil)

	// Set up tickets
	tickets := []domain.Ticket{
		{ID: "bb-1", Title: "Apple Pie", Status: "open", Priority: 1},
		{ID: "bb-2", Title: "Banana Split", Status: "open", Priority: 2},
		{ID: "bb-3", Title: "Cherry Tart", Status: "open", Priority: 3},
	}
	m.ticketList = newTicketList(tickets)

	// Apply filter matching all
	m.ticketList.SetFilterText("")

	// Select "Cherry Tart" (index 2)
	m.ticketList.Select(2)
	selectedItem, _ := m.ticketList.SelectedItem().(ticketItem)
	m.selection.Ticket = selectedItem.ticket

	// Refresh tickets where order changes but IDs stay same
	newTickets := []domain.Ticket{
		{ID: "bb-2", Title: "Banana Split", Status: "open", Priority: 2},
		{ID: "bb-3", Title: "Cherry Tart", Status: "open", Priority: 3},
		{ID: "bb-1", Title: "Apple Pie", Status: "open", Priority: 1},
	}
	msg := ticketsLoadedMsg(newTickets)
	updatedModel, _ := m.handleTicketsLoaded(msg)
	updatedM := updatedModel.(UIModel)

	// Selection should be preserved by finding ticket ID
	assert.Equal(t, "bb-3", updatedM.selection.Ticket.ID)
	assert.Equal(t, "Cherry Tart", updatedM.selection.Ticket.Title)

	// Filter should be preserved
	assert.Equal(t, "", updatedM.ticketList.FilterValue())
}

func TestHandleTicketsLoaded_PreservesPaginationWithFilter(t *testing.T) {
	app := newTestApp()
	app.ActiveProject = "."
	app.Stores = make(map[string]data.TicketStore)
	app.Stores["."] = &mockStore{}

	m := NewUIModel(app, nil)

	// Set up enough tickets to test pagination (10 items)
	tickets := make([]domain.Ticket, 10)
	for i := 0; i < 10; i++ {
		tickets[i] = domain.Ticket{
			ID:       fmt.Sprintf("bb-%d", i+1),
			Title:    fmt.Sprintf("Ticket %d", i+1),
			Status:   "open",
			Priority: i + 1,
		}
	}
	m.ticketList = newTicketList(tickets)

	// Apply filter matching all items
	m.ticketList.SetFilterText("Ticket")
	m.ticketList.SetFilterState(list.FilterApplied)

	// Select item at index 5 (would be on page 2 with typical page size)
	m.ticketList.Select(5)
	selectedItem, _ := m.ticketList.SelectedItem().(ticketItem)
	m.selection.Ticket = selectedItem.ticket

	// Set up window size to force pagination (small height)
	m.ticketList.SetSize(50, 20) // Width 50, Height 20

	// Refresh tickets
	msg := ticketsLoadedMsg(tickets)
	updatedModel, _ := m.handleTicketsLoaded(msg)
	updatedM := updatedModel.(UIModel)

	// Selection should be preserved
	assert.Equal(t, "bb-6", updatedM.selection.Ticket.ID)
	assert.Equal(t, "Ticket 6", updatedM.selection.Ticket.Title)

	// List cursor should be at same index
	newSelectedItem, ok := updatedM.ticketList.SelectedItem().(ticketItem)
	assert.True(t, ok)
	assert.Equal(t, "bb-6", newSelectedItem.ticket.ID)
	assert.Equal(t, 5, updatedM.ticketList.Index())

	// Filter should be preserved
	assert.Equal(t, list.FilterApplied, updatedM.ticketList.FilterState())
	assert.Equal(t, "Ticket", updatedM.ticketList.FilterValue())
}

func TestHandleTicketsLoaded_FilterByID(t *testing.T) {
	app := newTestApp()
	app.ActiveProject = "."
	app.Stores = make(map[string]data.TicketStore)
	app.Stores["."] = &mockStore{}

	m := NewUIModel(app, nil)

	// Set up tickets with different IDs
	tickets := []domain.Ticket{
		{ID: "bb-123", Title: "Fix login bug", Status: "open", Priority: 1},
		{ID: "bb-456", Title: "Add user profile", Status: "open", Priority: 2},
		{ID: "bb-789", Title: "Implement search", Status: "open", Priority: 3},
	}
	m.ticketList = newTicketList(tickets)

	// Apply filter that matches ID "123"
	m.ticketList.SetFilterText("123")
	m.ticketList.SetFilterState(list.FilterApplied)

	// Should find only the ticket with ID bb-123
	visibleItems := m.ticketList.VisibleItems()
	assert.Equal(t, 1, len(visibleItems))

	selectedItem, ok := visibleItems[0].(ticketItem)
	assert.True(t, ok)
	assert.Equal(t, "bb-123", selectedItem.ticket.ID)
	assert.Equal(t, "Fix login bug", selectedItem.ticket.Title)
}

func TestHandleTicketsLoaded_FilterByIDWithHyphen(t *testing.T) {
	app := newTestApp()
	app.ActiveProject = "."
	app.Stores = make(map[string]data.TicketStore)
	app.Stores["."] = &mockStore{}

	m := NewUIModel(app, nil)

	// Set up tickets with different IDs
	tickets := []domain.Ticket{
		{ID: "bb-123", Title: "Fix login bug", Status: "open", Priority: 1},
		{ID: "bb-456", Title: "Add user profile", Status: "open", Priority: 2},
		{ID: "bb-789", Title: "Implement search", Status: "open", Priority: 3},
	}
	m.ticketList = newTicketList(tickets)

	// Apply filter that matches full ID "bb-123"
	m.ticketList.SetFilterText("bb-123")
	m.ticketList.SetFilterState(list.FilterApplied)

	// Should find only the ticket with ID bb-123
	visibleItems := m.ticketList.VisibleItems()
	assert.Equal(t, 1, len(visibleItems))

	selectedItem, ok := visibleItems[0].(ticketItem)
	assert.True(t, ok)
	assert.Equal(t, "bb-123", selectedItem.ticket.ID)
	assert.Equal(t, "Fix login bug", selectedItem.ticket.Title)
}
