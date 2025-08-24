package main

import (
	"context"
	"log/slog"
	// "strings"
)

type EchoAgent struct {
	logger  *slog.Logger
	bus     *OptimizedBus
	command *Command
}

func (a *EchoAgent) Name() string { return "echo" }
func (a *EchoAgent) Start(ctx context.Context) error {
	ch, unsub, err := a.bus.Subscribe(EvtMessage, 64)
	if err != nil {
		return err
	}
	defer unsub()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case evt, ok := <-ch:
			if !ok {
				return nil
			}
			msg, _ := evt.Data.(MessageModel)

			switch msg.Type {
			case System:
				//
			case Human:
				if ok, _ := a.command.IsCommand(msg.Text); !ok {
					message := MessageModel{Type: Assistant, Text: "Echo Human: " + msg.Text}
					a.bus.Publish(EvtMessage, message)
				}
			case Assistant:
				//
			}
		}
	}
}
