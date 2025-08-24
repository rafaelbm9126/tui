package main

import (
	"log/slog"
	"strings"
)

type Command struct {
	logger *slog.Logger
	config *Config
	bus    *OptimizedBus
}

func NewCommand(logger *slog.Logger, config *Config, bus *OptimizedBus) *Command {
	return &Command{
		logger: logger,
		config: config,
		bus:    bus,
	}
}

func (c *Command) IsCommand(text string) bool {
	cmd := strings.TrimSpace(text)
	if !strings.HasPrefix(cmd, "/") {
		return false
	}

	// Procesa comandos
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}

	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	c.Execute(cmdName, args)
	return true
}

func (c *Command) Execute(cmd string, args []string) {

	_ = args

	switch cmd {
	case "help":
		c.bus.Publish(EvtSystem, c.config.Text.En.Comand.Help)
	default:
		c.logger.Error("Unknown command", "command", cmd)
	}
}
