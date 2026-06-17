// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package tmux

import (
	"context"
	"fmt"
	"strconv"

	"github.com/megatherium/blunderbust/internal/exec"
)

// defaultScrollback is the number of history lines capture-pane pulls in
// addition to the visible pane. 0 reproduces the legacy behaviour (visible
// pane only). A non-zero default restores useful agent history that was
// previously lost once output scrolled out of the viewport.
const defaultScrollback = 500

// OutputCapture reads a tmux pane via capture-pane and implements
// exec.OutputCapture so callers depend on the launcher-agnostic interface.
type OutputCapture struct {
	runner     CommandRunner
	windowID   string
	scrollback int
}

// NewOutputCapture creates a capture for the given window with the default
// scrollback range. The runner must not be nil for ReadOutput to succeed.
func NewOutputCapture(runner CommandRunner, windowID string) *OutputCapture {
	return &OutputCapture{
		runner:     runner,
		windowID:   windowID,
		scrollback: defaultScrollback,
	}
}

// NewOutputCaptureWithScrollback creates a capture with an explicit scrollback
// range (number of history lines). Use 0 to capture only the visible pane.
func NewOutputCaptureWithScrollback(runner CommandRunner, windowID string, scrollback int) *OutputCapture {
	return &OutputCapture{
		runner:     runner,
		windowID:   windowID,
		scrollback: scrollback,
	}
}

// Start begins capturing output from the tmux window (no-op since capture-pane
// is invoked directly on each ReadOutput).
func (c *OutputCapture) Start(ctx context.Context) (string, error) {
	return "", nil
}

// Stop ends the output capture (no-op).
func (c *OutputCapture) Stop(ctx context.Context) error {
	return nil
}

// ReadOutput captures the current content of the tmux pane, respecting ctx
// (previously hard-wired to context.Background()). The scrollback range is
// applied via -S so agent history that scrolled out of view is preserved.
func (c *OutputCapture) ReadOutput(ctx context.Context) ([]byte, error) {
	if c.windowID == "" {
		return nil, fmt.Errorf("window string is empty")
	}

	args := []string{"capture-pane", "-p", "-t", c.windowID}
	if c.scrollback > 0 {
		args = append(args, "-S", "-"+strconv.Itoa(c.scrollback))
	}

	out, err := c.runner.Run(ctx, "tmux", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to capture pane: %w", err)
	}

	return out, nil
}

// FilePath returns an empty string since capture streams directly from tmux.
func (c *OutputCapture) FilePath() string {
	return ""
}

// CaptureFactory builds OutputCapture instances for a shared runner. It
// implements exec.CaptureFactory so the App can hand out captures without
// importing the tmux package.
type CaptureFactory struct {
	runner CommandRunner
}

// NewCaptureFactory creates a factory that produces tmux OutputCaptures using
// the given runner.
func NewCaptureFactory(runner CommandRunner) *CaptureFactory {
	return &CaptureFactory{runner: runner}
}

// NewCapture returns an exec.OutputCapture bound to the given window, or nil
// when the window identifier is empty (no session to capture from).
func (f *CaptureFactory) NewCapture(launcherID string) exec.OutputCapture {
	if f == nil || launcherID == "" {
		return nil
	}
	return NewOutputCapture(f.runner, launcherID)
}

// Verify interface compliance at compile time.
var _ exec.OutputCapture = (*OutputCapture)(nil)
var _ exec.CaptureFactory = (*CaptureFactory)(nil)
