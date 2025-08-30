package agentspkg

import (
	"context"
	"log/slog"

	eventpkg "main/src/event"
	modelpkg "main/src/model"
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
			case modelpkg.TySystem:
				//
			case modelpkg.TyText:
				if ok, _ := a.Command.IsCommand(msg.Text); !ok {
					message := MessageModel{
						/**
						 * TODO: add thread_id
						 */
						ThreadId:  "",
						Type:      modelpkg.TyText,
						Source:    modelpkg.ScAssistant,
						WrittenBy: a.Name(),
						Text:      "Echo Human: " + msg.Text,
					}
					a.Bus.Publish(eventpkg.EvtMessage, message)
				}
			case modelpkg.TyCommand:
				//
			}
		}
	}
}
