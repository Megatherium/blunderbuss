// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package dolt

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTestBeadsIndicator(t *testing.T, beadsDir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(beadsDir, "config.yaml"), []byte("types:\n  custom: task\n"), 0644); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
}

func installMockBd(t *testing.T, tmpDir, script string) {
	t.Helper()
	mockBd := filepath.Join(tmpDir, "bd")
	if err := os.WriteFile(mockBd, []byte(script), 0755); err != nil {
		t.Fatalf("write mock bd: %v", err)
	}
	origPath := os.Getenv("PATH")
	t.Cleanup(func() { _ = os.Setenv("PATH", origPath) })
	if err := os.Setenv("PATH", tmpDir+":"+origPath); err != nil {
		t.Fatalf("set PATH: %v", err)
	}
}

func TestIsBeadsProject(t *testing.T) {
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	if err := os.MkdirAll(beadsDir, 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if IsBeadsProject(beadsDir) {
		t.Fatal("expected empty .beads to be false")
	}

	writeTestBeadsIndicator(t, beadsDir)
	if !IsBeadsProject(beadsDir) {
		t.Fatal("expected .beads with config.yaml to be true")
	}
}

func TestResolveConnection_ViaBdShowJSON(t *testing.T) {
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	if err := os.MkdirAll(beadsDir, 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeTestBeadsIndicator(t, beadsDir)

	installMockBd(t, tmpDir, `#!/bin/sh
if [ "$1" = "dolt" ] && [ "$2" = "show" ] && [ "$3" = "--json" ]; then
  echo '{"backend":"dolt","connection_ok":true,"database":"beads_bb","host":"127.0.0.1","port":34815,"shared_server":false,"user":"root"}'
  exit 0
fi
echo "unknown: $*" >&2
exit 1
`)

	meta, err := ResolveConnection(beadsDir)
	if err != nil {
		t.Fatalf("ResolveConnection: %v", err)
	}

	if meta.DoltDatabase != "beads_bb" {
		t.Errorf("database = %q, want beads_bb", meta.DoltDatabase)
	}
	if meta.ServerPort != 34815 {
		t.Errorf("port = %d, want 34815", meta.ServerPort)
	}
	if meta.ServerHost != "127.0.0.1" {
		t.Errorf("host = %q, want 127.0.0.1", meta.ServerHost)
	}
}

func TestResolveConnection_ConfigAndPortFile(t *testing.T) {
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	if err := os.MkdirAll(beadsDir, 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	configYAML := `issue-prefix: "bb"
dolt:
  host: 127.0.0.1
  port: 3307
  user: root
`
	if err := os.WriteFile(filepath.Join(beadsDir, "config.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(beadsDir, "dolt-server.port"), []byte("41234"), 0600); err != nil {
		t.Fatalf("write port file: %v", err)
	}

	installMockBd(t, tmpDir, `#!/bin/sh
exit 1
`)

	meta, err := ResolveConnection(beadsDir)
	if err != nil {
		t.Fatalf("ResolveConnection: %v", err)
	}

	if meta.DoltDatabase != "beads_bb" {
		t.Errorf("database = %q, want beads_bb", meta.DoltDatabase)
	}
	if meta.ServerPort != 41234 {
		t.Errorf("port = %d, want 41234 from dolt-server.port", meta.ServerPort)
	}
}

func TestResolveConnection_EnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	if err := os.MkdirAll(beadsDir, 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	metadataJSON := `{
		"backend": "dolt",
		"dolt_database": "beads_env",
		"dolt_server_host": "10.0.0.1",
		"dolt_server_port": 13307,
		"dolt_server_user": "meta-user"
	}`
	if err := os.WriteFile(filepath.Join(beadsDir, "metadata.json"), []byte(metadataJSON), 0644); err != nil {
		t.Fatalf("write metadata.json: %v", err)
	}

	installMockBd(t, tmpDir, `#!/bin/sh
exit 1
`)

	t.Setenv("BEADS_DOLT_SERVER_HOST", "192.168.1.10")
	t.Setenv("BEADS_DOLT_SERVER_PORT", "44001")
	t.Setenv("BEADS_DOLT_SERVER_USER", "env-user")
	t.Setenv("BEADS_DOLT_SERVER_DATABASE", "beads_override")

	meta, err := ResolveConnection(beadsDir)
	if err != nil {
		t.Fatalf("ResolveConnection: %v", err)
	}

	if meta.ServerHost != "192.168.1.10" {
		t.Errorf("host = %q", meta.ServerHost)
	}
	if meta.ServerPort != 44001 {
		t.Errorf("port = %d", meta.ServerPort)
	}
	if meta.ServerUser != "env-user" {
		t.Errorf("user = %q", meta.ServerUser)
	}
	if meta.DoltDatabase != "beads_override" {
		t.Errorf("database = %q", meta.DoltDatabase)
	}
}

func TestResolveConnection_SharedServerPortFile(t *testing.T) {
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	if err := os.MkdirAll(beadsDir, 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	configYAML := `dolt:
  shared-server: true
`
	if err := os.WriteFile(filepath.Join(beadsDir, "config.yaml"), []byte(configYAML), 0644); err != nil {
		t.Fatalf("write config.yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(beadsDir, "metadata.json"), []byte(`{"dolt_database":"beads_shared"}`), 0644); err != nil {
		t.Fatalf("write metadata.json: %v", err)
	}

	sharedDir := filepath.Join(tmpDir, "shared-server")
	if err := os.MkdirAll(sharedDir, 0750); err != nil {
		t.Fatalf("mkdir shared: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sharedDir, "dolt-server.port"), []byte("33081"), 0600); err != nil {
		t.Fatalf("write shared port: %v", err)
	}

	installMockBd(t, tmpDir, `#!/bin/sh
exit 1
`)
	t.Setenv("BEADS_SHARED_SERVER_DIR", sharedDir)

	meta, err := ResolveConnection(beadsDir)
	if err != nil {
		t.Fatalf("ResolveConnection: %v", err)
	}

	if meta.ServerPort != 33081 {
		t.Errorf("port = %d, want 33081 from shared-server port file", meta.ServerPort)
	}
}

func TestLookupCredentialsPassword(t *testing.T) {
	tmpDir := t.TempDir()
	credPath := filepath.Join(tmpDir, "credentials")
	content := `[127.0.0.1:3307]
password=local-secret

[10.0.0.5:4400]
password=remote-secret
`
	if err := os.WriteFile(credPath, []byte(content), 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}

	t.Setenv("BEADS_CREDENTIALS_FILE", credPath)
	t.Setenv("BEADS_DOLT_PASSWORD", "")

	if got := lookupCredentialsPassword("127.0.0.1", 3307); got != "local-secret" {
		t.Errorf("got %q, want local-secret", got)
	}
	if got := lookupCredentialsPassword("10.0.0.5", 4400); got != "remote-secret" {
		t.Errorf("got %q, want remote-secret", got)
	}
}

func TestResolveConnection_UninitializedWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	beadsDir := filepath.Join(tmpDir, ".beads")
	if err := os.MkdirAll(beadsDir, 0750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	_, err := ResolveConnection(beadsDir)
	if err == nil {
		t.Fatal("expected error for uninitialized workspace")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("unexpected error: %v", err)
	}
}