package commandpkg

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	buspkg "main/src/bus"
	configpkg "main/src/config"
	databasepkg "main/src/database"
	eventpkg "main/src/event"
	managerpkg "main/src/manager"
	messagepkg "main/src/message"
	modelpkg "main/src/model"
	toolspkg "main/src/tools"
)

type MessageModel = modelpkg.MessageModel
type ThreadModel = modelpkg.ThreadModel

type Command struct {
	logger   *slog.Logger
	config   *configpkg.Config
	bus      *buspkg.OptimizedBus
	db       *databasepkg.Database
	mgr      *managerpkg.Manager
	messages *messagepkg.MessageList
}

func NewCommand(
	logger *slog.Logger,
	config *configpkg.Config,
	bus *buspkg.OptimizedBus,
	db *databasepkg.Database,
	mgr *managerpkg.Manager,
	messages *messagepkg.MessageList,
) *Command {
	return &Command{
		logger:   logger,
		config:   config,
		bus:      bus,
		db:       db,
		mgr:      mgr,
		messages: messages,
	}
}

func (c *Command) IsCommandThenRun(text string) (bool, bool) {
	isCmd, parts := c.IsCommand(text)
	if !isCmd || len(parts) == 0 {
		// no command - show message //
		return false, true
	}

	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	save := c.Execute(cmdName, args)
	// command - show message?? //
	return true, save
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

func (c *Command) Execute(cmd string, args []string) bool {
	message := MessageModel{
		Type:   modelpkg.TySystem,
		Source: modelpkg.ScSystem,
	}

	switch cmd {
	case "q":
		c.bus.Publish(eventpkg.EvtSystem, "quit")

	case "h":
		commands := c.config.Config.Messages.Commands
		list := [][]string{}
		for _, item := range commands.Collection {
			list = append(list, []string{item.Command, item.Description})
			for _, variant := range item.Variants {
				list = append(list, []string{variant.Command, variant.Description})
			}
		}
		message.Text = commands.Title + "\n"
		message.Text += toolspkg.TableStatGeneral([]string{"Command", "Description"}, list)
		c.bus.Publish(eventpkg.EvtMessage, message)

	case "c":
		c.messages.Messages = []MessageModel{}
		// no show command //
		return false

	case "th":
		return ThreadCommand(c, args)

	case "st":
		agents := c.mgr.ListAgents()

		if len(agents) == 0 {
			message.Text = "No hay agentes registrados"
			c.bus.Publish(eventpkg.EvtMessage, message)
			break
		}

		list := [][]string{}
		for idx, agent := range agents {
			state := agent.State
			if agent.Restarts > 0 {
				state += fmt.Sprintf(" (%d reinicios)", agent.Restarts)
			}
			if agent.LastErr != nil {
				state += " ⚠️ " + agent.LastErr.Error()
			}
			list = append(list, []string{strconv.Itoa(idx + 1), agent.Name, state})
		}
		message.Text = "# Lista de agentes\n"
		message.Text += toolspkg.TableStatGeneral([]string{"#", "Nombre", "Estado"}, list)

		c.bus.Publish(eventpkg.EvtMessage, message)

	case "start":
		if len(args) < 1 {
			message.Text = "Uso: /start <agente>"
			c.bus.Publish(eventpkg.EvtMessage, message)
			break
		}
		name := args[0]
		err := c.mgr.StartAgent(name)
		if err != nil {
			message.Text = "Error al iniciar " + name + ": " + err.Error()
		} else {
			message.Text = "Iniciando " + name
		}
		c.bus.Publish(eventpkg.EvtMessage, message)

	case "stop":
		if len(args) < 1 {
			message.Text = "Uso: /stop <agente>"
			c.bus.Publish(eventpkg.EvtMessage, message)
			break
		}
		name := args[0]
		err := c.mgr.StopAgent(name)
		if err != nil {
			message.Text = "Error al detener " + name + ": " + err.Error()
		} else {
			message.Text = "Deteniendo " + name
		}
		c.bus.Publish(eventpkg.EvtMessage, message)

	case "restart":
		if len(args) < 1 {
			message.Text = "Uso: /restart <agente>"
			c.bus.Publish(eventpkg.EvtMessage, message)
			break
		}
		name := args[0]
		err := c.mgr.RestartAgent(name)
		if err != nil {
			message.Text = "Error al reiniciar " + name + ": " + err.Error()
		} else {
			message.Text = "Reiniciando " + name
		}

	default:
		message.Text = "**Command not found**"
		c.bus.Publish(eventpkg.EvtMessage, message)
	}

	// show command //
	return true
}
