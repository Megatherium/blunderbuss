// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package config

import "github.com/megatherium/blunderbust/internal/domain"

// Loader abstracts configuration file reading and validation.
type Loader interface {
	Load(path string) (*domain.Config, error)
	Save(path string, cfg *domain.Config) error
}

// TUILoader abstracts TUI-specific configuration I/O using the same Loader
// pattern as the main config. This lets the UI depend on an interface rather
// than calling package-level free functions, and makes TUI config testable
// with fakes. YAMLLoader implements both Loader and TUILoader.
type TUILoader interface {
	LoadTUI(path string) (*TUIConfig, error)
	SaveTUI(path string, cfg *TUIConfig) error
}
