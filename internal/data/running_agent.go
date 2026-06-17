// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package data

import (
	"context"
	"time"

	"github.com/megatherium/blunderbust/internal/domain"
)

// RunningAgentStore persists running-agent sessions so they survive restarts.
// Both the real backend (dolt) and the demo fake implement it, which lets the
// UI load/save/validate agents without type-asserting the concrete store.
type RunningAgentStore interface {
	UpsertRunningAgent(ctx context.Context, a domain.PersistedRunningAgent) error
	ListRunningAgentsByProjects(ctx context.Context, projectDirs []string) ([]domain.PersistedRunningAgent, error)
	ValidateAndPruneRunningAgents(ctx context.Context, projectDirs []string, inspector ProcessInspector) ([]domain.PersistedRunningAgent, error)
	DeleteStaleRunningAgents(ctx context.Context, maxAge time.Duration) error
}

// ProcessInspector probes the host for live processes. It is a data-layer
// abstraction (previously owned by the dolt package) so the validation/prune
// logic is testable independently of the real backend and its database.
type ProcessInspector interface {
	PIDExists(pid int) bool
	CommandForPID(ctx context.Context, pid int) (string, error)
}

// ConnectionRetryer is optionally implemented by stores that can recover from
// a lost server connection (e.g. by restarting the Dolt sql-server). The UI
// uses it to decide whether to show the [s]tart option and to issue a retry,
// instead of type-asserting the concrete *dolt.Store.
type ConnectionRetryer interface {
	CanRetryConnection() bool
	TryStartServer(ctx context.Context) (TicketStore, error)
}
