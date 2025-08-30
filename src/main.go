package main

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	agentspkg "main/src/agents"
	buspkg "main/src/bus"
	commandpkg "main/src/command"
	configpkg "main/src/config"
	databasepkg "main/src/database"
	eventpkg "main/src/event"
	managerpkg "main/src/manager"
	messagepkg "main/src/message"
	tuipkg "main/src/tui"
)

type MessageModel = messagepkg.MessageModel

func main() {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	conf := configpkg.LoadConfig()

	rootCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithCancel(rootCtx)

	db, _ := databasepkg.NewDatabase(logger)
	db.Migration()
	defer db.Close()

	bus := buspkg.NewMemoryBus(logger)
	defer bus.Close()

	mgr := managerpkg.NewManager(ctx, logger)

	messages := messagepkg.NewMessageList(db)

	command := commandpkg.NewCommand(logger, conf, bus, db, mgr, messages)

	tui := tuipkg.NewTUI(conf, bus, messages, command, logger)

	ev_sy, unsub_sy, err_sy := bus.Subscribe(eventpkg.EvtSystem, 64)
	go bus.RuntimeCaller(tui.Program(), ev_sy, err_sy)
	defer unsub_sy()

	ev_ms, unsub_ms, err_ms := bus.Subscribe(eventpkg.EvtMessage, 64)
	go bus.RuntimeCaller(tui.Program(), ev_ms, err_ms)
	defer unsub_ms()

	mgr.Register(&agentspkg.EchoAgent{Logger: logger, Bus: bus, Command: command}, true)
	mgr.Register(&agentspkg.AAgent{Logger: logger, Bus: bus, Command: command}, true)
	defer mgr.StopAll()

	if _, err := tui.Run(ctx, cancel); err != nil {
		logger.Error("Error starting TUI program", "error", err)
	}

	defer func() {
		os.Stdout.Write(buf.Bytes())
	}()
}
