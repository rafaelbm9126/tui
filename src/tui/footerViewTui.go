package tuipkg

import toolspkg "main/src/tools"

func FooterViewTui(t *TUI) string {
	text_footer_static := "ESC/Ctrl+C: Salir • PgUp/PgDn: Desplazar • ↑/↓: Historial"
	if t.showAlert {
		text_footer_static = toolspkg.SpaceBetween(
			t.viewport.Width,
			6,
			text_footer_static,
			t.spinner.View()+t.textAlert,
		)
	}

	return t.styles.help.
		Margin(0, 3, 0, 3).
		Render(text_footer_static)
}
