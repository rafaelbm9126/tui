package toolspkg

import (
	"strings"

	lipgloss "github.com/charmbracelet/lipgloss"
)

func SpaceBetween(width int, adjust int, left string, right string) string {
	space_between := (width - adjust) - lipgloss.Width(left) - lipgloss.Width(right)
	text := ""
	if space_between > 0 {
		text = strings.Repeat(" ", space_between)
	}
	return left + text + right
}
