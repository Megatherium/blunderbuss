package ui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/megatherium/blunderbust/internal/domain"
)

// AgentConfig holds configuration for rendering the agent output view
type AgentConfig struct {
	Agent  *RunningAgent
	Width  int
	Height int
	Theme  ThemePalette
}

// RenderAgentOutput renders the agent output view
func RenderAgentOutput(cfg AgentConfig) string {
	if cfg.Agent == nil {
		return "Agent not found\n\n[Press back to return]"
	}

	statusStr, statusColor := getAgentStatus(cfg.Agent.Info.Status)

	statusStyle := lipgloss.NewStyle().Foreground(statusColor).Bold(true)
	headerStyle := lipgloss.NewStyle().Bold(true).Underline(true)

	header := headerStyle.Render(fmt.Sprintf("Agent: %s", cfg.Agent.Info.Name))
	statusLine := fmt.Sprintf("Status: %s", statusStyle.Render(statusStr))
	launcherLine := fmt.Sprintf("Launcher: %s", cfg.Agent.Info.LauncherID)

	outputContent := getAgentOutputContent(cfg.Agent)

	outputStyle := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(ThemeInactive).
		Width(cfg.Width-4).
		Height(cfg.Height-10).
		Padding(0, 1)

	content := lipgloss.JoinVertical(lipgloss.Top,
		header,
		statusLine,
		launcherLine,
		"",
		"Output:",
		outputStyle.Render(outputContent),
		"",
		"[Press Enter to return to matrix]",
	)

	return content
}

func getAgentStatus(status domain.AgentStatus) (string, lipgloss.Color) {
	switch status {
	case domain.AgentRunning:
		return "Running", lipgloss.Color("34")
	case domain.AgentCompleted:
		return "Completed", lipgloss.Color("245")
	case domain.AgentFailed:
		return "Failed", lipgloss.Color("9")
	default:
		return "Unknown", lipgloss.Color("245")
	}
}

func getAgentOutputContent(agent *RunningAgent) string {
	if agent.LastOutput != "" {
		return agent.LastOutput
	}
	if agent.Info.Status == domain.AgentRunning {
		return "Waiting for output..."
	}
	return "No output available"
}
