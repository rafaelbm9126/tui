package toolspkg

import (
	"strings"
)

func FormatTableRows(rows [][]string, prefix string, suffix string) []string {
	if len(rows) == 0 {
		return []string{}
	}

	// 1️⃣ calcular ancho máximo de cada columna
	numCols := 0
	for _, r := range rows {
		if len(r) > numCols {
			numCols = len(r)
		}
	}

	widths := make([]int, numCols)
	for _, r := range rows {
		for i, col := range r {
			if len(col) > widths[i] {
				widths[i] = len(col)
			}
		}
	}

	// 2️⃣ construir filas alineadas
	var formattedRows []string
	for _, r := range rows {
		var sb strings.Builder
		for i, col := range r {
			sb.WriteString(col)
			spaces := widths[i] - len(col) + 2 // +2 para separación
			sb.WriteString(strings.Repeat(" ", spaces))
		}
		formattedRows = append(formattedRows, prefix+sb.String()+suffix)
	}

	return formattedRows
}
