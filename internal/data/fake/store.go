// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fake

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/megatherium/blunderbust/internal/data"
	"github.com/megatherium/blunderbust/internal/domain"
)

// errInvalidRunningAgent mirrors the validation the real backend performs on
// UpsertRunningAgent (required project/worktree/pid/harness fields).
var errInvalidRunningAgent = errors.New("invalid running agent data")

// TicketStore is an in-memory fake implementing data.TicketStore.
type TicketStore struct {
	mu      sync.RWMutex
	Tickets []domain.Ticket

	// RunningAgents is the in-memory running-agent store used to round-trip
	// persisted sessions in demo mode (data.RunningAgentStore). Each entry is
	// keyed by (project_dir, worktree_path, pid) to mirror the dolt unique key.
	runningAgents []domain.PersistedRunningAgent
	nextRunningID int
}

// Verify interface compliance at compile time.
var (
	_ data.TicketStore       = (*TicketStore)(nil)
	_ data.RunningAgentStore = (*TicketStore)(nil)
)

// ListTickets returns tickets matching the given filter.
func (s *TicketStore) ListTickets(_ context.Context, filter data.TicketFilter) ([]domain.Ticket, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []domain.Ticket
	for i := range s.Tickets {
		t := &s.Tickets[i]
		if filter.Status != "" && t.Status != filter.Status {
			continue
		}
		if filter.IssueType != "" && t.IssueType != filter.IssueType {
			continue
		}
		if filter.Search != "" && !strings.Contains(strings.ToLower(t.Title), strings.ToLower(filter.Search)) {
			continue
		}
		results = append(results, *t)
		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}
	return results, nil
}

// LatestUpdate returns the maximum updated_at timestamp from the ticket collection.
// Returns a zero time.Time if no tickets exist.
func (s *TicketStore) LatestUpdate(_ context.Context) (time.Time, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var latest time.Time
	for _, t := range s.Tickets {
		if t.UpdatedAt.After(latest) {
			latest = t.UpdatedAt
		}
	}
	return latest, nil
}

// NewWithSampleData returns a FakeTicketStore pre-loaded with sample tickets.
func NewWithSampleData() *TicketStore {
	now := time.Now()
	return &TicketStore{
		Tickets: []domain.Ticket{
			{ID: "bb-001", Title: "Bootstrap Go module", Status: "closed", Priority: 1, IssueType: "task", CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour)},
			{ID: "bb-002", Title: "Define core domain types", Status: "open", Priority: 1, IssueType: "task", CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now},
			{ID: "bb-003", Title: "Implement TicketStore backend", Status: "open", Priority: 1, IssueType: "task", CreatedAt: now.Add(-12 * time.Hour), UpdatedAt: now},
			{ID: "bb-004", Title: "Build TUI skeleton", Status: "open", Priority: 1, IssueType: "feature", CreatedAt: now.Add(-6 * time.Hour), UpdatedAt: now},
			{ID: "bb-005", Title: "Implement tmux launcher", Status: "open", Priority: 2, IssueType: "task", CreatedAt: now.Add(-3 * time.Hour), UpdatedAt: now},
		},
	}
}

// agentKey reports true if a persisted row matches the upsert unique key.
func agentMatchesKey(a domain.PersistedRunningAgent, projectDir, worktreePath string, pid int) bool {
	return a.ProjectDir == projectDir && a.WorktreePath == worktreePath && a.PID == pid
}

// UpsertRunningAgent inserts or updates a running-agent row in memory.
func (s *TicketStore) UpsertRunningAgent(_ context.Context, a domain.PersistedRunningAgent) error {
	if a.ProjectDir == "" || a.WorktreePath == "" || a.PID <= 0 || a.HarnessName == "" {
		return errInvalidRunningAgent
	}
	if a.LauncherID == "" {
		a.LauncherID = "unknown"
	}
	now := time.Now()
	a.LastSeen = now

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.runningAgents {
		if agentMatchesKey(s.runningAgents[i], a.ProjectDir, a.WorktreePath, a.PID) {
			a.ID = s.runningAgents[i].ID
			a.StartedAt = s.runningAgents[i].StartedAt
			s.runningAgents[i] = a
			return nil
		}
	}
	s.nextRunningID++
	a.ID = s.nextRunningID
	a.StartedAt = now
	s.runningAgents = append(s.runningAgents, a)
	return nil
}

// ListRunningAgentsByProjects returns running agents whose project_dir matches,
// newest first, mirroring the dolt ordering.
func (s *TicketStore) ListRunningAgentsByProjects(_ context.Context, projectDirs []string) ([]domain.PersistedRunningAgent, error) {
	if len(projectDirs) == 0 {
		return nil, nil
	}
	wanted := make(map[string]struct{}, len(projectDirs))
	for _, d := range projectDirs {
		wanted[d] = struct{}{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.PersistedRunningAgent, 0, len(s.runningAgents))
	for _, a := range s.runningAgents {
		if _, ok := wanted[a.ProjectDir]; ok {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out, nil
}

// ValidateAndPruneRunningAgents accepts every persisted agent as valid in demo
// mode (there is no real OS process to probe). The inspector, if provided, is
// consulted so the path is exercised identically to the real backend.
func (s *TicketStore) ValidateAndPruneRunningAgents(ctx context.Context, projectDirs []string, inspector data.ProcessInspector) ([]domain.PersistedRunningAgent, error) {
	agents, err := s.ListRunningAgentsByProjects(ctx, projectDirs)
	if err != nil {
		return nil, err
	}
	valid := make([]domain.PersistedRunningAgent, 0, len(agents))
	for _, a := range agents {
		if inspector != nil && !inspector.PIDExists(a.PID) {
			s.deleteRunningAgentByKey(a.ProjectDir, a.WorktreePath, a.PID)
			continue
		}
		valid = append(valid, a)
	}
	return valid, nil
}

// DeleteStaleRunningAgents removes rows whose last_seen is older than maxAge.
func (s *TicketStore) DeleteStaleRunningAgents(_ context.Context, maxAge time.Duration) error {
	if maxAge <= 0 {
		maxAge = time.Hour
	}
	cutoff := time.Now().Add(-maxAge)
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := s.runningAgents[:0]
	for _, a := range s.runningAgents {
		if a.LastSeen.After(cutoff) {
			kept = append(kept, a)
		}
	}
	s.runningAgents = kept
	return nil
}

func (s *TicketStore) deleteRunningAgentByKey(projectDir, worktreePath string, pid int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := s.runningAgents[:0]
	for _, a := range s.runningAgents {
		if !agentMatchesKey(a, projectDir, worktreePath, pid) {
			kept = append(kept, a)
		}
	}
	s.runningAgents = kept
}
