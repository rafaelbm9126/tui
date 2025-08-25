package agentspkg

import (
	"context"
	"log/slog"

	eventpkg "main/src/event"
	messagepkg "main/src/message"
)

type EchoAgent struct {
	Logger  *slog.Logger
	Bus     *OptimizedBus
	Command *Command
}

func (a *EchoAgent) Name() string { return "echo" }

func (a *EchoAgent) Start(ctx context.Context) error {
	ch, unsub, err := a.Bus.Subscribe(eventpkg.EvtMessage, 64)
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
			case messagepkg.System:
				//
			case messagepkg.Human:
				if ok, _ := a.Command.IsCommand(msg.Text); !ok {
					message := MessageModel{
						Type: messagepkg.Assistant,
						From: a.Name(),
						Text: "Echo Human: " + msg.Text,
					}
					a.Bus.Publish(eventpkg.EvtMessage, message)
				}
			case messagepkg.Assistant:
				//
			}
		}
	}
}
