package main

import (
	"bytes"
	"context"
	"log/slog"
	"main/src/tui"
	"os"
	"os/signal"
	"syscall"

	"main/src/agents"
	"main/src/bus"
	"main/src/command"
	"main/src/config"
	"main/src/event"
	"main/src/manager"
	"main/src/message"
)

type MessageModel = messagepkg.MessageModel

func main() {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	conf := configpkg.LoadConfig()
	_ = conf

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithCancel(rootCtx)

	_ = cancel

	bus := buspkg.NewMemoryBus(logger)
	defer bus.Close()

	mgr := managerpkg.NewManager(ctx, logger)

	command := commandpkg.NewCommand(logger, conf, bus, mgr)

	_ = command

	messages := messagepkg.NewMessageList()

	tui := tuipkg.NewTUI(bus, messages, command, logger)

	ev_sy, unsub_sy, err_sy := bus.Subscribe(eventpkg.EvtSystem, 64)
	go bus.RuntimeCaller(tui.Program(), ev_sy, err_sy)
	defer unsub_sy()

	ev_ms, unsub_ms, err_ms := bus.Subscribe(eventpkg.EvtMessage, 64)
	go bus.RuntimeCaller(tui.Program(), ev_ms, err_ms)
	defer unsub_ms()

	mgr.Register(&agentspkg.EchoAgent{Logger: logger, Bus: bus, Command: command}, true)
	mgr.Register(&agentspkg.AAgent{Logger: logger, Bus: bus, Command: command}, true)
	// mgr.StartAll()
	defer mgr.StopAll()

	bus.Publish(eventpkg.EvtMessage, MessageModel{Type: messagepkg.System, Text: "Hello World..!"})

	if _, err := tui.Run(ctx, cancel); err != nil {
		logger.Error("Error starting TUI program", "error", err)
	}

	defer func() {
		os.Stdout.Write(buf.Bytes())
	}()
}
