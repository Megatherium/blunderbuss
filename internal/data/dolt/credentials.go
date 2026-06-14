// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package dolt

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const credentialsPasswordEnvVar = "BEADS_DOLT_PASSWORD"

// defaultCredentialsPath returns the beads credentials file location.
func defaultCredentialsPath() string {
	if p := os.Getenv("BEADS_CREDENTIALS_FILE"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "beads", "credentials")
}

// resolveServerPassword returns the Dolt server password using beads priority:
// 1. BEADS_DOLT_PASSWORD env var
// 2. credentials file [host:port] section
func resolveServerPassword(host string, port int) string {
	if p := os.Getenv(credentialsPasswordEnvVar); p != "" {
		return p
	}
	return lookupCredentialsPassword(host, port)
}

// lookupCredentialsPassword reads the INI-style beads credentials file.
func lookupCredentialsPassword(host string, port int) string {
	path := defaultCredentialsPath()
	if path == "" {
		return ""
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	section := fmt.Sprintf("[%s:%d]", host, port)
	inSection := false

	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inSection = strings.EqualFold(line, section)
			continue
		}

		if !inSection {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(key), "password") {
			return strings.TrimSpace(value)
		}
	}

	return ""
}
