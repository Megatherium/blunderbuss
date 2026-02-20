// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package dolt

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Mode represents the Dolt connection mode.
type Mode string

const (
	// EmbeddedMode uses github.com/dolthub/driver (CGO required).
	// Database is stored locally in .beads/dolt/
	EmbeddedMode Mode = "embedded"
	// ServerMode connects to a running dolt sql-server via MySQL protocol.
	ServerMode Mode = "server"
)

// Metadata represents the parsed .beads/metadata.json file.
type Metadata struct {
	// Database backend type (should be "dolt")
	Backend string `json:"backend"`
	// DoltDatabase is the database name within Dolt (e.g., "beads_bb")
	DoltDatabase string `json:"dolt_database"`
	// DoltMode indicates whether to use embedded or server mode
	DoltMode string `json:"dolt_mode"`
	// ServerHost is the hostname for server mode connections
	ServerHost string `json:"dolt_server_host"`
	// ServerPort is the port for server mode connections
	ServerPort int `json:"dolt_server_port"`
	// ServerUser is the MySQL user for server mode connections
	ServerUser string `json:"dolt_server_user"`
}

// ConnectionMode determines the connection mode from the metadata.
// Returns ServerMode if dolt_mode is "server" or if server connection
// fields are present. Otherwise returns EmbeddedMode.
func (m *Metadata) ConnectionMode() Mode {
	if m.DoltMode == "server" {
		return ServerMode
	}
	// Also detect server mode by presence of server fields
	if m.ServerHost != "" && m.ServerPort > 0 {
		return ServerMode
	}
	return EmbeddedMode
}

// IsValid returns true if the metadata contains the minimum required fields.
func (m *Metadata) IsValid() bool {
	return m.DoltDatabase != ""
}

// LoadMetadata reads and parses the metadata.json file from the given beads directory.
// Returns actionable errors for common failure scenarios.
func LoadMetadata(beadsDir string) (*Metadata, error) {
	metadataPath := filepath.Join(beadsDir, "metadata.json")

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"no beads database found at %q: metadata.json is missing\n"+
					"Is this a beads project? Run 'bd init' to initialize beads in this repository",
				beadsDir,
			)
		}
		return nil, fmt.Errorf("failed to read metadata.json: %w", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf(
			"metadata.json is corrupted or has invalid JSON: %w\n"+
				"Try removing %s and running 'bd init' to recreate it",
			err, metadataPath,
		)
	}

	if !metadata.IsValid() {
		return nil, fmt.Errorf(
			"metadata.json is missing required field 'dolt_database'\n"+
				"File location: %s\n"+
				"Try running 'bd init' to regenerate the metadata file",
			metadataPath,
		)
	}

	return &metadata, nil
}

// DoltDir returns the path to the Dolt database directory.
// This is always beadsDir/dolt for embedded mode.
func DoltDir(beadsDir string) string {
	return filepath.Join(beadsDir, "dolt")
}
