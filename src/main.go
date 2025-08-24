package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	// "time"
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

	tui := NewTUI(bus, messages, logger)

	es, unsub, err := bus.Subscribe(0, 64)
	if err != nil {
		return
	}
	go func() {
		for evt := range es {
			logger.Info("ES Received event:", "Data", evt.Data)
			msg := Event{evt: evt}
			tui.Program().Send(msg)
		}
	}()
	defer unsub()

	em, unsub2, err := bus.Subscribe(1, 64)
	if err != nil {
		return
	}
	go func() {
		for evt := range em {
			logger.Info("EM Received event")
			msg := Event{evt: evt}
			tui.Program().Send(msg)
		}
	}()
	defer unsub2()

	bus.Publish(0, "Default Main [0]")
	bus.Publish(1, MessageModel{Type: Assistant, Text: "Default Main [1]"})

	if _, err := tui.Run(ctx, cancel); err != nil {
		logger.Error("Error starting TUI program", "error", err)
	}
}
