package toolspkg

import (
	"fmt"
	"strings"
)

func TableStatGeneral(headers []string, rows [][]string) string {
	var sb strings.Builder
	maxLenCol := make([]int, len(headers))

	for _, row := range rows {
		for idxCol, mlc := range maxLenCol {
			if len(row[idxCol]) > mlc {
				maxLenCol[idxCol] = len(row[idxCol])
			}
		}
	}

	// encabezado
	sb.WriteString("| ")
	for idxh, header := range headers {
		sb.WriteString(
			fmt.Sprintf(
				"%-*s |",
				maxLenCol[idxh],
				header,
			),
		)
	}
	sb.WriteString("\n|")
	for mlc, _ := range maxLenCol {
		sb.WriteString(
			fmt.Sprintf(
				"-%s-|",
				strings.Repeat("-", mlc),
			),
		)
	}
	sb.WriteString("\n")

	// filas
	sb.WriteString("|")
	for _, row := range rows {
		for idxc, col := range row {
			sb.WriteString(
				fmt.Sprintf(
					" %-*s |",
					maxLenCol[idxc],
					col,
				),
			)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
