// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Package exec provides launcher-agnostic execution abstractions for launching
// and monitoring development harnesses.
//
// This package defines the interfaces every launcher backend implements. The
// concrete tmux backend lives in internal/exec/tmux; future backends (docker,
// headless os/exec, ...) implement the same interfaces so the App and UI never
// import a launcher-specific package.
//
// The interfaces:
//
//   - Launcher launches a harness session from a LaunchSpec and returns a
//     LaunchResult.
//   - StatusChecker reports whether a previously launched session is still
//     alive, returning an AgentStatus (StatusRunning/StatusDead/StatusUnknown).
//   - OutputCapture reads the live output of a launched session. All methods
//     take a context so captures honour cancellation.
//   - CaptureFactory builds an OutputCapture for a given launcher id; the App
//     hands captures out through it without importing a backend package.
//
// Adding a new launcher backend (e.g. docker) is therefore confined to
// internal/exec/<backend>/ plus a single wiring change at the composition root
// (cmd/blunderbust) and (if capture routing differs) the App.NewCapture type
// switch.
package exec
