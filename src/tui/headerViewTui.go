package tuipkg

import (
	"fmt"
	modelpkg "main/src/model"

	"github.com/charmbracelet/lipgloss"
)

type ThreadModel = modelpkg.ThreadModel

// top, right, bottom, left //

func HeaderViewTui(t *TUI) string {
	threadName := "<empty>"

	if t.messages.Thread != nil {
		threadName = t.messages.Thread.Name
	}

	return t.styles.header.
		Margin(0, 0, 0, 4).
		Render("AATUI") + t.styles.help.
		Margin(0, 0, 0, 2).
		Foreground(lipgloss.Color("#29BEB0")).
		Render(fmt.Sprintf(
			"[Thread | %d | %s ]",
			len(t.messages.Threads),
			threadName,
		))
}
