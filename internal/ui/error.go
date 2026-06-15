package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/megatherium/blunderbust/internal/data/dolt"
)

var (
	errorTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			MarginBottom(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6666"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			MarginTop(2)
)

// errorView renders an error screen with actionable messaging.
func errorView(err error, hasRetry, hasStart bool) string {
	if err == nil {
		return "An unknown error occurred"
	}

	errStr := err.Error()
	var b strings.Builder

	b.WriteString(errorTitleStyle.Render("Error"))
	b.WriteString("\n\n")

	// Provide user-friendly messages based on error type.
	// Prefer typed Is* checks (from dolt package) over raw err.Error() substrings
	// for connection and server errors (per bb-970u.2). Fall back to strings only
	// for high-level workspace messages that aren't yet typed.
	switch {
	case strings.Contains(errStr, "no beads workspace found"),
		strings.Contains(errStr, "beads workspace at"),
		strings.Contains(errStr, "metadata.json is missing"):
		b.WriteString(errorStyle.Render("No beads database found."))
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(errStr))
		b.WriteString("\n\n")
		b.WriteString("Is this a beads project? Run 'bd init' or 'bd bootstrap' to initialize beads in this repository.")

	case strings.Contains(errStr, "dolt database directory not found"):
		b.WriteString(errorStyle.Render("The beads database is not initialized."))
		b.WriteString("\n\n")
		b.WriteString("Run 'bd init' to create the beads database.")

	case dolt.IsErrServerNotRunning(err):
		b.WriteString(errorStyle.Render(errStr))

	case dolt.IsConnectionError(err):
		b.WriteString(errorStyle.Render("Cannot connect to Dolt server."))
		b.WriteString("\n\n")
		b.WriteString("Please check that the Dolt server is running and the connection details are correct.")

	default:
		// Show the original error for unknown cases. Avoid new direct strings.Contains
		// on connection-related errs here.
		b.WriteString(errorStyle.Render(errStr))
	}

	b.WriteString("\n")

	var helpParts []string
	if hasRetry {
		helpParts = append(helpParts, "[r] Retry connection")
	}
	if hasStart {
		helpParts = append(helpParts, "[s] Start server")
	}
	helpParts = append(helpParts, "[q] Quit")

	if len(helpParts) > 0 {
		b.WriteString(helpStyle.Render("\n" + strings.Join(helpParts, "  ")))
	} else {
		b.WriteString(helpStyle.Render("\nPress 'q' to quit."))
	}

	return b.String()
}
