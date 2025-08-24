package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithCancel(rootCtx)

	messages := NewMessageList()
	messages.AddMessageSystem("# Hello")
	messages.AddMessageHuman("Hello..!")
	messages.AddMessageAssistant("Hello..!")

	bus := NewMemoryBus(logger)
	defer bus.Close()

	es, unsub, err := bus.Subscribe(EvtSystem, 64)
	if err != nil {
		return
	}
	defer unsub()

	em, unsub2, err := bus.Subscribe(EvtMessage, 64)
	if err != nil {
		return
	}
	defer unsub2()

	bus.Publish(EvtSystem, "Default")

	tui := NewTUI(bus, messages, logger)

	p, err := tui.Run(ctx, cancel)
	if err != nil {
		logger.Error("Error starting TUI program", "error", err)
	}

	go func() {
		for evt := range es {
			logger.Info("ES Received event:", "Data", evt.Data)
			msg := Event{evt: evt}
			p.Send(msg)
		}
		for evt := range em {
			logger.Info("EM Received event")
			msg := Event{evt: evt}
			p.Send(msg)
		}
	}()

	time.Sleep(100 * time.Millisecond)
}
