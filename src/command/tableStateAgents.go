package commandpkg

import (
	"fmt"
	"strings"

	managerpkg "main/src/manager"
)

func TableStateAgents(agents []managerpkg.AgentStatus) string {
	var sb strings.Builder

	// calcular anchos máximos
	maxName := len("Name")
	maxState := len("State")

	for _, ag := range agents {
		if len(ag.Name) > maxName {
			maxName = len(ag.Name)
		}
		state := ag.State
		if ag.Restarts > 0 {
			state += fmt.Sprintf(" (%d reinicios)", ag.Restarts)
		}
		if ag.LastErr != nil {
			state += " ⚠️ " + ag.LastErr.Error()
		}
		if len(state) > maxState {
			maxState = len(state)
		}
	}

	// encabezado
	sb.WriteString(fmt.Sprintf("| %-5s | %-*s | %-*s |\n", "Index", maxName, "Name", maxState, "State"))
	sb.WriteString(fmt.Sprintf("|-------|-%s-|-%s-|\n", strings.Repeat("-", maxName), strings.Repeat("-", maxState)))

	// filas
	for i, ag := range agents {
		state := ag.State
		if ag.Restarts > 0 {
			state += fmt.Sprintf(" (%d reinicios)", ag.Restarts)
		}
		if ag.LastErr != nil {
			state += " ⚠️ " + ag.LastErr.Error()
		}
		sb.WriteString(fmt.Sprintf("| %-5d | %-*s | %-*s |\n", i+1, maxName, ag.Name, maxState, state))
	}

	return sb.String()
}
