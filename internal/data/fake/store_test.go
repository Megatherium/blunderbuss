// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fake

import (
	"context"
	"testing"
	"time"

	"github.com/megatherium/blunderbust/internal/data"
	"github.com/megatherium/blunderbust/internal/domain"
)

func TestFakeStore_ListTickets_All(t *testing.T) {
	now := time.Now()
	store := &TicketStore{
		Tickets: []domain.Ticket{
			{ID: "bb-001", Title: "First", Status: "open", Priority: 1, IssueType: "task", CreatedAt: now, UpdatedAt: now},
			{ID: "bb-002", Title: "Second", Status: "closed", Priority: 2, IssueType: "bug", CreatedAt: now, UpdatedAt: now},
		},
	}

	filter := data.TicketFilter{}
	tickets, err := store.ListTickets(context.Background(), filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tickets) != 2 {
		t.Errorf("expected 2 tickets, got %d", len(tickets))
	}
}

func TestFakeStore_ListTickets_WithStatusFilter(t *testing.T) {
	now := time.Now()
	store := &TicketStore{
		Tickets: []domain.Ticket{
			{ID: "bb-001", Title: "First", Status: "open", Priority: 1, IssueType: "task", CreatedAt: now, UpdatedAt: now},
			{ID: "bb-002", Title: "Second", Status: "closed", Priority: 2, IssueType: "bug", CreatedAt: now, UpdatedAt: now},
		},
	}

	filter := data.TicketFilter{Status: "open"}
	tickets, err := store.ListTickets(context.Background(), filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tickets) != 1 {
		t.Errorf("expected 1 ticket, got %d", len(tickets))
	}

	if tickets[0].ID != "bb-001" {
		t.Errorf("expected bb-001, got %s", tickets[0].ID)
	}
}

func TestFakeStore_ListTickets_WithSearchFilter(t *testing.T) {
	now := time.Now()
	store := &TicketStore{
		Tickets: []domain.Ticket{
			{ID: "bb-001", Title: "Test ticket one", Status: "open", Priority: 1, IssueType: "task", CreatedAt: now, UpdatedAt: now},
			{ID: "bb-002", Title: "Another ticket", Status: "open", Priority: 2, IssueType: "bug", CreatedAt: now, UpdatedAt: now},
		},
	}

	filter := data.TicketFilter{Search: "test"}
	tickets, err := store.ListTickets(context.Background(), filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tickets) != 1 {
		t.Errorf("expected 1 ticket, got %d", len(tickets))
	}

	if tickets[0].ID != "bb-001" {
		t.Errorf("expected bb-001, got %s", tickets[0].ID)
	}
}

func TestFakeStore_ListTickets_WithLimit(t *testing.T) {
	now := time.Now()
	store := &TicketStore{
		Tickets: []domain.Ticket{
			{ID: "bb-001", Title: "First", Status: "open", Priority: 1, IssueType: "task", CreatedAt: now, UpdatedAt: now},
			{ID: "bb-002", Title: "Second", Status: "open", Priority: 2, IssueType: "bug", CreatedAt: now, UpdatedAt: now},
			{ID: "bb-003", Title: "Third", Status: "open", Priority: 3, IssueType: "task", CreatedAt: now, UpdatedAt: now},
		},
	}

	filter := data.TicketFilter{Limit: 2}
	tickets, err := store.ListTickets(context.Background(), filter)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tickets) != 2 {
		t.Errorf("expected 2 tickets, got %d", len(tickets))
	}
}

func TestFakeStore_LatestUpdate_HasTickets(t *testing.T) {
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	weekAgo := now.Add(-7 * 24 * time.Hour)

	store := &TicketStore{
		Tickets: []domain.Ticket{
			{ID: "bb-001", Title: "Old", Status: "open", Priority: 1, IssueType: "task", CreatedAt: weekAgo, UpdatedAt: weekAgo},
			{ID: "bb-002", Title: "Medium", Status: "open", Priority: 2, IssueType: "bug", CreatedAt: yesterday, UpdatedAt: yesterday},
			{ID: "bb-003", Title: "New", Status: "open", Priority: 3, IssueType: "task", CreatedAt: now, UpdatedAt: now},
		},
	}

	latest, err := store.LatestUpdate(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !latest.Equal(now) {
		t.Errorf("expected %v, got %v", now, latest)
	}
}

func TestFakeStore_LatestUpdate_Empty(t *testing.T) {
	store := &TicketStore{
		Tickets: []domain.Ticket{},
	}

	latest, err := store.LatestUpdate(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !latest.IsZero() {
		t.Errorf("expected zero time, got %v", latest)
	}
}
