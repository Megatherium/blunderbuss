// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package dolt

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// Mode represents the Dolt connection mode.
type Mode string

const (
	// ServerMode connects to a running dolt sql-server via MySQL protocol.
	ServerMode Mode = "server"
)

// Metadata represents resolved beads Dolt connection settings.
type Metadata struct {
	// Database backend type (should be "dolt")
	Backend string `json:"backend"`
	// DoltDatabase is the database name within Dolt (e.g., "beads_bb")
	DoltDatabase string `json:"dolt_database"`
	// DoltMode indicates whether to use server mode (always "server")
	DoltMode string `json:"dolt_mode"`
	// ServerHost is the hostname for server mode connections
	ServerHost string `json:"dolt_server_host"`
	// ServerPort is the port for server mode connections
	ServerPort int `json:"dolt_server_port"`
	// ServerUser is the MySQL user for server mode connections
	ServerUser string `json:"dolt_server_user"`
	// ServerReadyTimeoutSeconds is the timeout in seconds to wait for Dolt server to be ready
	ServerReadyTimeoutSeconds int `json:"dolt_server_ready_timeout"`
}

// ConnectionMode always returns ServerMode since embedded mode is not supported.
func (m *Metadata) ConnectionMode() Mode {
	return ServerMode
}

// IsValid returns true if the metadata contains the minimum required fields.
func (m *Metadata) IsValid() bool {
	return m.DoltDatabase != ""
}

// ServerReadyTimeout returns the configured server ready timeout or default 10 seconds.
func (m *Metadata) ServerReadyTimeout() time.Duration {
	if m.ServerReadyTimeoutSeconds > 0 {
		return time.Duration(m.ServerReadyTimeoutSeconds) * time.Second
	}
	return 10 * time.Second
}

// ResolveServerPort fills in ServerPort when it is unset, using the same sources
// as beads: env vars, dolt-server.port, config.yaml, then bd dolt status.
func (m *Metadata) ResolveServerPort(ctx context.Context, beadsDir string) (int, error) {
	if m.ServerPort > 0 {
		return m.ServerPort, nil
	}

	yamlCfg, yamlErr := loadBeadsYAMLConfig(beadsDir)
	if yamlErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", filepath.Join(beadsDir, "config.yaml"), yamlErr)
	}
	resolveRuntimePort(beadsDir, m, yamlCfg)
	if m.ServerPort > 0 {
		return m.ServerPort, nil
	}

	port, err := detectPortFromDoltStatus(ctx, beadsDir)
	if err != nil {
		return 0, fmt.Errorf("failed to detect Dolt server port: %w", err)
	}
	if port > 0 {
		m.ServerPort = port
		return port, nil
	}

	return 0, nil
}

// detectPortFromDoltStatus runs 'bd dolt status' and extracts the port from output.
func detectPortFromDoltStatus(ctx context.Context, beadsDir string) (int, error) {
	projectDir := filepath.Dir(beadsDir)

	cmdCtx, cancel := context.WithTimeout(ctx, bdCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "bd", "dolt", "status")
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, nil //nolint:nilerr // Expected when dolt is not initialized
	}

	outputStr := string(output)
	patterns := []string{
		`(?:^|\s)Port:\s*(\d+)`,
		`Expected port:\s*(\d+)`,
	}

	for _, pattern := range patterns {
		portRegex := regexp.MustCompile(pattern)
		matches := portRegex.FindStringSubmatch(outputStr)

		if len(matches) >= 2 {
			port, err := strconv.Atoi(matches[1])
			if err != nil {
				return 0, fmt.Errorf("failed to parse port number from 'bd dolt status': %w", err)
			}
			if port >= 1024 && port <= 65535 {
				return port, nil
			}
		}
	}

	return 0, nil
}

// DoltDir returns the path to the Dolt database directory.
func DoltDir(beadsDir string) string {
	return filepath.Join(beadsDir, "dolt")
}
