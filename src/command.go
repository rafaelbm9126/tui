package main

import (
	"fmt"
	"log/slog"
	"strings"
)

type Command struct {
	logger *slog.Logger
	config *Config
	bus    *OptimizedBus
	mgr    *Manager
}

func NewCommand(
	logger *slog.Logger,
	config *Config,
	bus *OptimizedBus,
	mgr *Manager,
) *Command {
	return &Command{
		logger: logger,
		config: config,
		bus:    bus,
		mgr:    mgr,
	}
}

func (c *Command) IsCommandThenRun(text string) bool {
	isCmd, parts := c.IsCommand(text)
	if !isCmd || len(parts) == 0 {
		return false
	}

	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	c.Execute(cmdName, args)
	return true
}

func (c *Command) IsCommand(text string) (bool, []string) {
	cmd := strings.TrimSpace(text)
	if !strings.HasPrefix(cmd, "/") {
		return false, nil
	}

	// Procesa comandos
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false, nil
	}

	return true, parts
}

func (c *Command) Execute(cmd string, args []string) {

	_ = args

	switch cmd {
	case "quit", "q":
		c.bus.Publish(EvtSystem, "quit")
	case "help", "h":
		message := MessageModel{Type: System, Text: c.config.Text["messages"]["commands"]["help"]}
		c.bus.Publish(EvtMessage, message)
	case "status", "st":
		agents := c.mgr.ListAgents()

		if len(agents) == 0 {
			message := MessageModel{Type: System, Text: "No hay agentes registrados"}
			c.bus.Publish(EvtMessage, message)
			break
		}
		var sb strings.Builder
		sb.WriteString("# Estado de agentes\n\n")
		for _, ag := range agents {
			sb.WriteString("- **")
			sb.WriteString(ag.Name)
			sb.WriteString("**: ")
			sb.WriteString(ag.State)
			if ag.Restarts > 0 {
				sb.WriteString(" (")
				sb.WriteString(fmt.Sprintf("%d", ag.Restarts))
				sb.WriteString(" reinicios)")
			}
			if ag.LastErr != nil {
				sb.WriteString("\n  _Error: ")
				sb.WriteString(ag.LastErr.Error())
				sb.WriteString("_")
			}
			sb.WriteString("\n")
		}

		message := MessageModel{Type: System, Text: sb.String()}
		c.bus.Publish(EvtMessage, message)
	default:
		message := MessageModel{Type: System, Text: "**Command not found**"}
		c.bus.Publish(EvtMessage, message)
	}
}
