// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package fake

import (
	"context"
	"sync"

	"github.com/megatherium/blunderbuss/internal/domain"
	"github.com/megatherium/blunderbuss/internal/exec"
)

// Launcher is a fake implementing exec.Launcher that captures launch
// attempts and returns predictable results without spawning real tmux windows.
type Launcher struct {
	mu       sync.Mutex
	Launches []domain.LaunchSpec
	Err      error
}

// Verify interface compliance at compile time.
var _ exec.Launcher = (*Launcher)(nil)

// Launch records the spec and returns a synthetic result.
func (l *Launcher) Launch(_ context.Context, spec domain.LaunchSpec) (*domain.LaunchResult, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.Launches = append(l.Launches, spec)

	if l.Err != nil {
		return &domain.LaunchResult{Error: l.Err}, l.Err
	}

	return &domain.LaunchResult{
		WindowName: spec.WindowName,
		WindowID:   "fake-window-0",
		PaneID:     "fake-pane-0",
	}, nil
}
