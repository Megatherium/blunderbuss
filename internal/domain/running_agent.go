package domain

import "time"

// LauncherType represents the type of launcher that started an agent.
type LauncherType int

const (
	LauncherTypeUnknown LauncherType = iota
	LauncherTypeTmux
	LauncherTypeDocker
)

// String returns the string representation of the launcher type.
func (t LauncherType) String() string {
	switch t {
	case LauncherTypeTmux:
		return "tmux"
	case LauncherTypeDocker:
		return "docker"
	default:
		return "unknown"
	}
}

// PersistedRunningAgent represents one row in the running_agents table.
type PersistedRunningAgent struct {
	ID            int
	ProjectDir    string
	WorktreePath  string
	PID           int
	LauncherType  LauncherType
	LauncherID    string
	Ticket        string
	TicketTitle   string
	HarnessName   string
	HarnessBinary string
	Model         string
	Agent         string
	StartedAt     time.Time
	LastSeen      time.Time
}
