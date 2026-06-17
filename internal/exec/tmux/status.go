// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tmux

import (
	"context"
	"strings"

	"github.com/megatherium/blunderbust/internal/exec"
)

// StatusChecker monitors tmux window liveness and implements exec.StatusChecker
// so callers depend on the launcher-agnostic interface.
type StatusChecker struct {
	runner CommandRunner
}

// NewStatusChecker creates a new StatusChecker backed by the given runner.
func NewStatusChecker(runner CommandRunner) *StatusChecker {
	return &StatusChecker{
		runner: runner,
	}
}

// CheckStatus determines if the tmux window named launcherID still exists.
// It uses `tmux list-windows -F '#{window_name} #{window_id}'` and matches on
// the window name (which is set to the launcher ID at launch time).
// Returns exec.StatusUnknown when the underlying tmux call fails, so a
// transient tmux error is not misreported as a dead agent.
func (c *StatusChecker) CheckStatus(ctx context.Context, launcherID string) exec.AgentStatus {
	output, err := c.runner.Run(ctx, "tmux", "list-windows", "-F", "#{window_name} #{window_id}")
	if err != nil {
		return exec.StatusUnknown
	}

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		if parts[0] == launcherID {
			return exec.StatusRunning
		}
	}

	return exec.StatusDead
}

// Verify interface compliance at compile time.
var _ exec.StatusChecker = (*StatusChecker)(nil)
