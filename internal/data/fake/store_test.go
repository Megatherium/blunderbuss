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

// TestFakeStore_RunningAgentStore_Interface confirms the fake satisfies the
// data.RunningAgentStore contract (demo-mode persistence).
func TestFakeStore_RunningAgentStore_Interface(t *testing.T) {
	var _ data.RunningAgentStore = (*TicketStore)(nil)
}

func TestFakeStore_RunningAgent_RoundTrip(t *testing.T) {
	store := &TicketStore{}
	ctx := context.Background()

	agent := domain.PersistedRunningAgent{
		ProjectDir:    "/repo",
		WorktreePath:  "/repo/wt",
		PID:           4242,
		LauncherType:  domain.LauncherTypeTmux,
		LauncherID:    "bb-1",
		Ticket:        "bb-1",
		TicketTitle:   "Demo agent",
		HarnessName:   "kilocode",
		HarnessBinary: "kilo",
		Model:         "m",
		Agent:         "a",
	}

	if err := store.UpsertRunningAgent(ctx, agent); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := store.ListRunningAgentsByProjects(ctx, []string{"/repo"})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 || got[0].PID != 4242 || got[0].LauncherType != domain.LauncherTypeTmux {
		t.Fatalf("unexpected list result: %+v", got)
	}
	if got[0].TicketTitle != "Demo agent" {
		t.Fatalf("ticket title not round-tripped: %q", got[0].TicketTitle)
	}

	// Upsert again with the same key -> update in place, no duplicate.
	agent.TicketTitle = "Updated"
	if err := store.UpsertRunningAgent(ctx, agent); err != nil {
		t.Fatalf("upsert(2): %v", err)
	}
	got, _ = store.ListRunningAgentsByProjects(ctx, []string{"/repo"})
	if len(got) != 1 {
		t.Fatalf("expected 1 agent after re-upsert, got %d", len(got))
	}
	if got[0].TicketTitle != "Updated" {
		t.Fatalf("expected updated title, got %q", got[0].TicketTitle)
	}

	// ValidateAndPrune with an inspector that reports the PID alive keeps it.
	valid, err := store.ValidateAndPruneRunningAgents(ctx, []string{"/repo"}, aliveInspector{4242: struct{}{}})
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if len(valid) != 1 {
		t.Fatalf("expected 1 valid agent, got %d", len(valid))
	}

	// An inspector reporting the PID dead prunes it.
	valid, err = store.ValidateAndPruneRunningAgents(ctx, []string{"/repo"}, aliveInspector{})
	if err != nil {
		t.Fatalf("validate(dead): %v", err)
	}
	if len(valid) != 0 {
		t.Fatalf("expected 0 valid agents after prune, got %d", len(valid))
	}
	got, _ = store.ListRunningAgentsByProjects(ctx, []string{"/repo"})
	if len(got) != 0 {
		t.Fatalf("pruned agent should be deleted, got %d", len(got))
	}
}

func TestFakeStore_RunningAgent_DeleteStale(t *testing.T) {
	store := &TicketStore{}
	ctx := context.Background()

	fresh := domain.PersistedRunningAgent{
		ProjectDir: "/repo", WorktreePath: "/repo", PID: 1, HarnessName: "kilo",
	}
	stale := domain.PersistedRunningAgent{
		ProjectDir: "/repo", WorktreePath: "/repo", PID: 2, HarnessName: "kilo",
	}
	if err := store.UpsertRunningAgent(ctx, fresh); err != nil {
		t.Fatalf("upsert fresh: %v", err)
	}
	if err := store.UpsertRunningAgent(ctx, stale); err != nil {
		t.Fatalf("upsert stale: %v", err)
	}

	// Force the stale agent's last_seen far into the past.
	store.mu.Lock()
	for i := range store.runningAgents {
		if store.runningAgents[i].PID == 2 {
			store.runningAgents[i].LastSeen = time.Now().Add(-2 * time.Hour)
		}
	}
	store.mu.Unlock()

	if err := store.DeleteStaleRunningAgents(ctx, time.Hour); err != nil {
		t.Fatalf("delete stale: %v", err)
	}
	got, _ := store.ListRunningAgentsByProjects(ctx, []string{"/repo"})
	if len(got) != 1 || got[0].PID != 1 {
		t.Fatalf("expected only fresh PID=1 to remain, got %+v", got)
	}
}

func TestFakeStore_RunningAgent_InvalidRejected(t *testing.T) {
	store := &TicketStore{}
	if err := store.UpsertRunningAgent(context.Background(), domain.PersistedRunningAgent{}); err == nil {
		t.Fatal("expected error for invalid agent, got nil")
	}
}

// aliveInspector implements data.ProcessInspector; PIDs in the set are alive.
type aliveInspector map[int]struct{}

func (a aliveInspector) PIDExists(pid int) bool { _, ok := a[pid]; return ok }
func (a aliveInspector) CommandForPID(_ context.Context, _ int) (string, error) {
	return "kilo", nil
}
