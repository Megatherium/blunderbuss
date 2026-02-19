// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package domain

import "time"

// TemplateContext is the fat context passed to both command and prompt
// templates. It is intentionally generous — templates pick what they need.
type TemplateContext struct {
	// Ticket fields
	TicketID          string
	TicketTitle       string
	TicketDescription string
	TicketStatus      string
	TicketPriority    int
	TicketIssueType   string
	TicketAssignee    string
	TicketCreatedAt   time.Time
	TicketUpdatedAt   time.Time

	// Harness fields
	HarnessName string

	// Selection fields
	Model string
	Agent string

	// Environment fields
	RepoPath string
	Branch   string
	WorkDir  string
	User     string
	Hostname string

	// Runtime fields
	DryRun    bool
	Debug     bool
	Timestamp time.Time
}
