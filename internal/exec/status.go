// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package exec

import "context"

// AgentStatus is the launcher-agnostic liveness of a launched agent session.
// It is reported by StatusChecker so callers (the UI) never need to import a
// concrete launcher package (e.g. internal/exec/tmux) to interpret status.
type AgentStatus int

const (
	// StatusUnknown indicates the status could not be determined (e.g. the
	// underlying command failed). Callers usually treat this as "still
	// running" rather than dead to avoid false negatives.
	StatusUnknown AgentStatus = iota
	// StatusRunning means the launched session is alive.
	StatusRunning
	// StatusDead means the launched session has terminated.
	StatusDead
)

// String returns a human-readable representation of the status.
func (s AgentStatus) String() string {
	switch s {
	case StatusRunning:
		return "Running"
	case StatusDead:
		return "Dead"
	case StatusUnknown:
		return "Unknown"
	default:
		return "Invalid"
	}
}

// StatusChecker reports whether a previously launched agent session is still
// alive. Each launcher backend (tmux, docker, ...) provides its own concrete
// implementation; consumers depend only on this interface.
type StatusChecker interface {
	CheckStatus(ctx context.Context, launcherID string) AgentStatus
}
