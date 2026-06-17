// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package exec

import "context"

// OutputCapture reads the live output of a launched agent session. Each
// launcher backend provides its own concrete implementation (e.g. tmux reads
// a pane, a future headless launcher tails a log file). Consumers depend only
// on this interface so the UI does not import launcher-specific packages.
//
// All methods accept a context so long-lived capture operations (e.g. reading
// a tmux pane) can be cancelled; this was previously hard-wired to
// context.Background().
type OutputCapture interface {
	// Start prepares the capture. The returned string is the backing file
	// path when applicable, or empty when the backend streams directly.
	Start(ctx context.Context) (string, error)
	// Stop releases any resources held by the capture.
	Stop(ctx context.Context) error
	// ReadOutput returns the current captured content.
	ReadOutput(ctx context.Context) ([]byte, error)
	// FilePath returns the backing file path, or empty when not file-backed.
	FilePath() string
}

// CaptureFactory builds an OutputCapture for a previously launched session
// identified by launcherID. It decouples capture creation from the concrete
// backend so the App can hand out captures without importing a launcher
// package. A nil capture (or a nil factory) signals that capture is
// unavailable for this session/launcher type.
type CaptureFactory interface {
	NewCapture(launcherID string) OutputCapture
}
