package tuipkg

import modelpkg "main/src/model"

type ThreadModel = modelpkg.ThreadModel

func HeaderViewTui(t *TUI) string {
	var threadName string

	if t.messages.Thread != nil {
		threadName = t.messages.Thread.Name
	}

	return t.styles.header.
		Margin(0, 0, 0, 2).
		Render(" AATUI ") + t.styles.help.
		Margin(0, 0, 0, 0).
		Render("Thread: "+threadName)
}
