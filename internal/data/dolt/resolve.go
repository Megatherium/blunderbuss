// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package dolt

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	defaultDoltServerHost = "127.0.0.1"
	defaultDoltServerPort = 3307
	defaultDoltServerUser = "root"
	defaultDoltDatabase   = "beads"
	defaultSharedPort     = 3308
	bdCommandTimeout      = 10 * time.Second
)

type doltShowJSON struct {
	Backend      string `json:"backend"`
	ConnectionOK bool   `json:"connection_ok"`
	Database     string `json:"database"`
	Host         string `json:"host"`
	Port         int    `json:"port"`
	SharedServer bool   `json:"shared_server"`
	User         string `json:"user"`
}

type beadsYAMLConfig struct {
	IssuePrefix string         `yaml:"issue-prefix"`
	Dolt        yamlDoltConfig `yaml:"dolt"`
}

type yamlDoltConfig struct {
	Mode         string `yaml:"mode"`
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	SharedServer bool   `yaml:"shared-server"`
}

// IsBeadsProject reports whether beadsDir looks like an initialized beads workspace.
func IsBeadsProject(beadsDir string) bool {
	info, err := os.Stat(beadsDir)
	if err != nil || !info.IsDir() {
		return false
	}

	indicators := []string{
		"config.yaml",
		"metadata.json",
		"issues.jsonl",
		filepath.Join("dolt", ".dolt"),
		filepath.Join("dolt", ".bd-dolt-ok"),
	}
	for _, name := range indicators {
		if _, err := os.Stat(filepath.Join(beadsDir, name)); err == nil {
			return true
		}
	}

	entries, err := os.ReadDir(filepath.Join(beadsDir, "dolt"))
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				return true
			}
		}
	}

	return false
}

// ResolveConnection resolves Dolt sql-server connection settings for a beads project.
// It delegates to "bd dolt show --json" when available, then falls back to a local
// layered resolver that mirrors beads' config precedence.
func ResolveConnection(ctx context.Context, beadsDir string) (*Metadata, error) {
	if err := validateBeadsDir(beadsDir); err != nil {
		return nil, err
	}

	projectDir := filepath.Dir(beadsDir)
	if meta, err := resolveViaBdShow(ctx, projectDir, beadsDir); err == nil {
		return meta, nil
	}
	// bd dolt show failed; fall back to local layered resolver (no warning spam;
	// actionable errors are returned from resolveLocal if it also fails).
	// Debug logging of the bd show err (if needed) is left to callers.

	meta, err := resolveLocal(beadsDir)
	if err != nil {
		return nil, err
	}

	if meta.DoltDatabase == "" {
		return nil, fmt.Errorf(
			"could not determine Dolt database name for %q\n"+
				"Try running 'bd init' or 'bd bootstrap' in this repository",
			beadsDir,
		)
	}

	return meta, nil
}

func validateBeadsDir(beadsDir string) error {
	info, err := os.Stat(beadsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(
				"no beads workspace found at %q\n"+
					"Is this a beads project? Run 'bd init' to initialize beads in this repository",
				beadsDir,
			)
		}
		return fmt.Errorf("cannot access beads directory %q: %w", beadsDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%q exists but is not a directory", beadsDir)
	}
	if !IsBeadsProject(beadsDir) {
		return fmt.Errorf(
			"beads workspace at %q is not initialized\n"+
				"Run 'bd init' or 'bd bootstrap' to initialize the database",
			beadsDir,
		)
	}
	return nil
}

func resolveViaBdShow(ctx context.Context, projectDir, beadsDir string) (*Metadata, error) {
	if _, err := exec.LookPath("bd"); err != nil {
		return nil, fmt.Errorf("bd not found in PATH")
	}

	cmdCtx, cancel := context.WithTimeout(ctx, bdCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "bd", "dolt", "show", "--json")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bd dolt show --json: %w: %s", err, strings.TrimSpace(string(output)))
	}

	var show doltShowJSON
	if err := json.Unmarshal(output, &show); err != nil {
		return nil, fmt.Errorf("parsing bd dolt show output: %w", err)
	}

	if show.Database == "" {
		return nil, fmt.Errorf("bd dolt show returned empty database name")
	}

	meta := &Metadata{
		Backend:      show.Backend,
		DoltDatabase: show.Database,
		DoltMode:     "server",
		ServerHost:   show.Host,
		ServerPort:   show.Port,
		ServerUser:   show.User,
	}
	if meta.ServerHost == "" {
		meta.ServerHost = defaultDoltServerHost
	}
	if meta.ServerUser == "" {
		meta.ServerUser = defaultDoltServerUser
	}

	fileMeta, _ := parseMetadataFile(beadsDir)
	if fileMeta != nil && fileMeta.ServerReadyTimeoutSeconds > 0 {
		meta.ServerReadyTimeoutSeconds = fileMeta.ServerReadyTimeoutSeconds
	}

	return meta, nil
}

func resolveLocal(beadsDir string) (*Metadata, error) {
	meta, err := parseMetadataFile(beadsDir)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		meta = &Metadata{Backend: "dolt", DoltMode: "server"}
	}

	yamlCfg, yamlErr := loadBeadsYAMLConfig(beadsDir)
	if yamlErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to parse %s: %v\n", filepath.Join(beadsDir, "config.yaml"), yamlErr)
	}
	applyYAMLDefaults(meta, yamlCfg)
	applyEnvOverrides(meta)
	resolveRuntimePort(beadsDir, meta, yamlCfg)

	if meta.ServerHost == "" {
		meta.ServerHost = defaultDoltServerHost
	}
	if meta.ServerUser == "" {
		meta.ServerUser = defaultDoltServerUser
	}
	if meta.DoltDatabase == "" {
		meta.DoltDatabase = inferDatabaseName(beadsDir, yamlCfg)
	}

	return meta, nil
}

func parseMetadataFile(beadsDir string) (*Metadata, error) {
	metadataPath := filepath.Join(beadsDir, "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
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

	return &metadata, nil
}

func loadBeadsYAMLConfig(beadsDir string) (*beadsYAMLConfig, error) {
	path := filepath.Join(beadsDir, "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cfg beadsYAMLConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config.yaml: %w", err)
	}
	return &cfg, nil
}

func applyYAMLDefaults(meta *Metadata, yamlCfg *beadsYAMLConfig) {
	if yamlCfg == nil {
		return
	}
	if meta.ServerHost == "" && yamlCfg.Dolt.Host != "" {
		meta.ServerHost = yamlCfg.Dolt.Host
	}
	if meta.ServerUser == "" && yamlCfg.Dolt.User != "" {
		meta.ServerUser = yamlCfg.Dolt.User
	}
	if meta.DoltMode == "" && yamlCfg.Dolt.Mode != "" {
		meta.DoltMode = yamlCfg.Dolt.Mode
	}
}

func applyEnvOverrides(meta *Metadata) {
	if host := os.Getenv("BEADS_DOLT_SERVER_HOST"); host != "" {
		meta.ServerHost = host
	}
	if port := envInt("BEADS_DOLT_SERVER_PORT", "BEADS_DOLT_PORT"); port > 0 {
		meta.ServerPort = port
	}
	if user := os.Getenv("BEADS_DOLT_SERVER_USER"); user != "" {
		meta.ServerUser = user
	}
	if db := os.Getenv("BEADS_DOLT_SERVER_DATABASE"); db != "" {
		meta.DoltDatabase = db
	}
}

func envInt(keys ...string) int {
	for _, key := range keys {
		if raw := strings.TrimSpace(os.Getenv(key)); raw != "" {
			if port, err := strconv.Atoi(raw); err == nil && port > 0 {
				return port
			}
		}
	}
	return 0
}

// readPortFile reads the runtime port from a dolt-server.port file.
func readPortFile(serverDir string) int {
	data, err := os.ReadFile(filepath.Join(serverDir, "dolt-server.port"))
	if err != nil {
		return 0
	}
	port, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || port <= 0 {
		return 0
	}
	return port
}

func resolveRuntimePort(beadsDir string, meta *Metadata, yamlCfg *beadsYAMLConfig) {
	// Env overrides are applied before this call and always win.
	if envInt("BEADS_DOLT_SERVER_PORT", "BEADS_DOLT_PORT") > 0 {
		return
	}

	serverDir := resolveServerDir(beadsDir, yamlCfg)
	if port := readPortFile(serverDir); port > 0 {
		meta.ServerPort = port
		return
	}

	if yamlCfg != nil && yamlCfg.Dolt.Port > 0 {
		meta.ServerPort = yamlCfg.Dolt.Port
		return
	}

	if meta.ServerPort > 0 {
		return
	}

	if isSharedServerMode(yamlCfg) {
		meta.ServerPort = defaultSharedPort
	}
}

func isSharedServerMode(yamlCfg *beadsYAMLConfig) bool {
	if v := os.Getenv("BEADS_DOLT_SHARED_SERVER"); v == "1" || strings.EqualFold(v, "true") {
		return true
	}
	return yamlCfg != nil && yamlCfg.Dolt.SharedServer
}

func resolveServerDir(beadsDir string, yamlCfg *beadsYAMLConfig) string {
	if !isSharedServerMode(yamlCfg) {
		return beadsDir
	}
	if custom := os.Getenv("BEADS_SHARED_SERVER_DIR"); custom != "" {
		return custom
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return beadsDir
	}
	return filepath.Join(home, ".beads", "shared-server")
}

func inferDatabaseName(beadsDir string, yamlCfg *beadsYAMLConfig) string {
	if yamlCfg != nil && yamlCfg.IssuePrefix != "" {
		return "beads_" + yamlCfg.IssuePrefix
	}

	doltDir := filepath.Join(beadsDir, "dolt")
	entries, err := os.ReadDir(doltDir)
	if err != nil {
		return defaultDoltDatabase
	}

	var candidates []string
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == ".dolt" {
			continue
		}
		if _, err := os.Stat(filepath.Join(doltDir, entry.Name(), ".dolt")); err == nil {
			candidates = append(candidates, entry.Name())
		}
	}
	if len(candidates) == 1 {
		return candidates[0]
	}

	return defaultDoltDatabase
}

// LoadMetadata resolves connection settings for a beads project.
// It is kept for backward compatibility; prefer ResolveConnection.
func LoadMetadata(beadsDir string) (*Metadata, error) {
	return ResolveConnection(context.Background(), beadsDir)
}
