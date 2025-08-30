package commandpkg

import (
	"strconv"
	"strings"
	"time"

	eventpkg "main/src/event"
	modelpkg "main/src/model"
	toolspkg "main/src/tools"
)

func ThreadCommand(c *Command, args []string) {
	message := MessageModel{
		Type:   modelpkg.TySystem,
		Source: modelpkg.ScSystem,
	}
	c.messages.Threads, _ = c.db.ListThreads()

	switch args[0] {
	case "-c":
		thread, _ := c.db.CreateThread(ThreadModel{
			Name:      args[1],
			CreatedAt: time.Now(),
		})
		message.Text = "Nuevo Thread [" + thread.Id + "] " + args[1]
	case "-l":
		list := [][]string{}
		for idx, thread := range c.messages.Threads {
			list = append(list, []string{strconv.Itoa(idx + 1), thread.Id, thread.Name})
		}
		message.Text = "# Lista de flags disponibles\n"
		message.Text += toolspkg.TableStatGeneral([]string{"#", "ID", "Nombre"}, list)

	case "-u":
		ok, idx := ThreadCommandValidation(3, args, c.messages.Threads)
		if !ok {
			message.Text = "Uso: `/th` -u [index] [name]"
			break
		}
		thread := c.messages.Threads[idx-1]
		thread.Name = strings.Join(args[2:], " ")
		c.db.UpdateThread(thread)
		message.Text = "Update Thread [" + thread.Id + "] " + args[1]
	case "-s":
		ok, idx := ThreadCommandValidation(2, args, c.messages.Threads)
		if !ok {
			message.Text = "Uso: `/th` -s [index]"
			break
		}
		thread := c.messages.Threads[idx-1]
		c.messages.Thread = &thread
		message.Text = "Select Thread [" + thread.Id + "] " + args[1]
	case "-d":
		ok, idx := ThreadCommandValidation(2, args, c.messages.Threads)
		if !ok {
			message.Text = "Uso: `/th` -d [index]"
			break
		}
		thread := c.messages.Threads[idx-1]
		c.db.DeleteThread(thread)
		message.Text = "Delete Thread [" + thread.Id + "] " + args[1]
	default:
		message.Text = "Flag no reconocido"
	}

	c.bus.Publish(eventpkg.EvtMessage, message)
}

func ThreadCommandValidation(arg int, args []string, Threads []ThreadModel) (bool, int) {
	if len(args) < arg {
		return false, 0
	}
	idx, err := strconv.Atoi(args[1])
	if err != nil {
		return false, 0
	}
	if idx < 1 || idx > len(Threads) {
		return false, 0
	}
	return true, idx
}
