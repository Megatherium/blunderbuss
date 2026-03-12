// Copyright (C) 2026 megatherium
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package domain

import "time"

// SidebarNodeType represents the type of a node in the sidebar tree.
type SidebarNodeType int

// Node type constants for sidebar nodes.
const (
	NodeTypeProject SidebarNodeType = iota
	NodeTypeWorktree
	NodeTypeHarness
	NodeTypeAgent
)

// String returns the string representation of the node type.
func (t SidebarNodeType) String() string {
	switch t {
	case NodeTypeProject:
		return "project"
	case NodeTypeWorktree:
		return "worktree"
	case NodeTypeHarness:
		return "harness"
	case NodeTypeAgent:
		return "agent"
	default:
		return "unknown"
	}
}

// AgentStatus represents the current status of an agent.
type AgentStatus int

const (
	AgentRunning AgentStatus = iota
	AgentCompleted
	AgentFailed
)

// String returns the string representation of the agent status.
func (s AgentStatus) String() string {
	switch s {
	case AgentRunning:
		return "running"
	case AgentCompleted:
		return "completed"
	case AgentFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// SidebarNode represents a node in the sidebar tree hierarchy.
// Nodes can be projects (containing worktrees), worktrees, harnesses, or agents.
type SidebarNode struct {
	ID            string
	Name          string
	Path          string
	Type          SidebarNodeType
	Children      []SidebarNode
	IsExpanded    bool
	IsRunning     bool
	ParentProject *SidebarNode

	WorktreeInfo *WorktreeInfo
	HarnessInfo  *HarnessInfo
	AgentInfo    *AgentInfo
}

// WorktreeInfo contains metadata about a git worktree.
type WorktreeInfo struct {
	Name       string
	Path       string
	Branch     string
	CommitHash string
	IsMain     bool
	IsDirty    bool
}

// HarnessInfo contains metadata about a running harness session.
type HarnessInfo struct {
	LauncherID string
	TicketID   string
	StartedAt  time.Time
	Status     string
}

// AgentInfo contains metadata about a running agent session.
type AgentInfo struct {
	ID           string
	Name         string
	LauncherID   string
	WorktreePath string
	Status       AgentStatus
	StartedAt    time.Time
	TicketID     string
	TicketTitle  string
	HarnessName  string
	ModelName    string
	AgentName    string
}
