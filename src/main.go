package main

import (
	"bytes"
	"context"
	"log/slog"
	"main/src/agents"
	"main/src/bus"
	"main/src/command"
	"main/src/config"
	"main/src/event"
	"main/src/logger"
	"main/src/manager"
	"main/src/message"
	"main/src/tui"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	config := LoadConfig()

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithCancel(rootCtx)

	bus := NewMemoryBus(logger)
	defer bus.Close()

	mgr := NewManager(ctx, logger)

	command := NewCommand(logger, config, bus, mgr)

	messages := NewMessageList(command)

	tui := NewTUI(bus, messages, logger)

	ev_sy, unsub_sy, err_sy := bus.Subscribe(EvtSystem, 64)
	go bus.RuntimeCaller(tui, ev_sy, err_sy)
	defer unsub_sy()

	ev_ms, unsub_ms, err_ms := bus.Subscribe(EvtMessage, 64)
	go bus.RuntimeCaller(tui, ev_ms, err_ms)
	defer unsub_ms()

	mgr.Register(&EchoAgent{logger: logger, bus: bus, command: command}, true)
	mgr.Register(&AAgent{logger: logger, bus: bus, command: command}, true)
	// mgr.StartAll()
	defer mgr.StopAll()

	bus.Publish(EvtMessage, MessageModel{Type: System, Text: "Hello World..!"})

	if _, err := tui.Run(ctx, cancel); err != nil {
		logger.Error("Error starting TUI program", "error", err)
	}

	defer func() {
		os.Stdout.Write(buf.Bytes())
	}()
}
