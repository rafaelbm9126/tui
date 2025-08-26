package tuipkg

func HeaderViewTui(t *TUI) string {
	return t.styles.header.
		Margin(0, 2, 0, 2).
		Render(" Agentes TUI ") + "  " + t.styles.help.
		Margin(0, 2, 0, 2).
		Render("(Enter: enviar · /help · /md on|off · ↑↓ historial · PgUp/PgDn scroll · Esc/Ctrl+C salir)")
}
