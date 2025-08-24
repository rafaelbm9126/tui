package main

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// =========================
// Contrato Agent y Manager dinámico
// =========================

// Agent define la interfaz para los agentes
type Agent interface {
	Name() string
	Start(ctx context.Context) error
}

// Manager gestiona el ciclo de vida de los agentes
type Manager struct {
	log *slog.Logger

	mu     sync.Mutex
	ctx    context.Context
	agents map[string]*runner
}

type runner struct {
	agent       Agent
	autoRestart bool
	minBackoff  time.Duration
	maxBackoff  time.Duration

	mu       sync.Mutex
	state    string // "stopped" | "running"
	restarts int
	lastErr  error

	parentCtx context.Context
	runCtx    context.Context
	cancel    context.CancelFunc
	done      chan struct{}
	stopping  bool
}

// NewManager crea un nuevo Manager de agentes
func NewManager(ctx context.Context, log *slog.Logger) *Manager {
	return &Manager{
		log:    log,
		ctx:    ctx,
		agents: make(map[string]*runner),
	}
}

// Register registra un agente en el Manager
func (m *Manager) Register(agent Agent, autoRestart bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := agent.Name()
	if m.agents == nil {
		m.agents = make(map[string]*runner)
	}
	if _, exists := m.agents[name]; exists {
		m.log.Warn("replacing existing agent", "name", name)
	}

	m.agents[name] = &runner{
		agent:       agent,
		autoRestart: autoRestart,
		minBackoff:  100 * time.Millisecond,
		maxBackoff:  5 * time.Second,
		state:       "stopped",
	}
}

// StartAgent inicia un agente específico
func (m *Manager) StartAgent(name string) error {
	m.mu.Lock()
	r, exists := m.agents[name]
	if !exists {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.state == "running" {
		return nil
	}
	if r.stopping {
		return nil
	}

	// Crea contexto para la ejecución del agente
	ctx, cancel := context.WithCancel(m.ctx)
	r.runCtx = ctx
	r.cancel = cancel
	r.parentCtx = m.ctx
	r.done = make(chan struct{})
	r.state = "running"

	m.log.Info("starting agent", "name", name)

	// Inicia el agente en una goroutine
	go func() {
		defer close(r.done)
		for {
			select {
			case <-r.parentCtx.Done():
				r.mu.Lock()
				r.state = "stopped"
				r.mu.Unlock()
				return
			default:
			}

			// Ejecuta el agente
			err := r.agent.Start(r.runCtx)

			r.mu.Lock()
			// Si el contexto fue cancelado, termina normalmente
			if err == context.Canceled || r.runCtx.Err() == context.Canceled {
				r.state = "stopped"
				r.mu.Unlock()
				return
			}

			// Si hay error y debe reiniciarse, lo intenta de nuevo con backoff
			if err != nil && r.autoRestart && !r.stopping {
				r.restarts++
				r.lastErr = err
				backoff := r.minBackoff * time.Duration(1<<(r.restarts-1))
				if backoff > r.maxBackoff {
					backoff = r.maxBackoff
				}
				m.log.Error("agent error, restarting", "name", name, "error", err, "backoff", backoff, "restarts", r.restarts)
				r.mu.Unlock()
				time.Sleep(backoff)
				continue
			}

			// Si no debe reiniciarse o se está deteniendo, termina
			r.state = "stopped"
			if err != nil {
				r.lastErr = err
				m.log.Error("agent terminated", "name", name, "error", err)
			} else {
				m.log.Info("agent terminated", "name", name)
			}
			r.mu.Unlock()
			return
		}
	}()

	return nil
}

// StopAgent detiene un agente específico
func (m *Manager) StopAgent(name string) error {
	m.mu.Lock()
	r, exists := m.agents[name]
	if !exists {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	r.mu.Lock()
	if r.state != "running" {
		r.mu.Unlock()
		return nil
	}
	r.stopping = true
	cancel := r.cancel
	done := r.done
	r.mu.Unlock()

	// Cancelar el contexto
	m.log.Info("stopping agent", "name", name)
	if cancel != nil {
		cancel()
	}

	// Esperar a que la goroutine termine
	select {
	case <-done:
		// OK, terminó normalmente
	case <-time.After(5 * time.Second):
		m.log.Warn("agent stop timeout", "name", name)
	}

	r.mu.Lock()
	r.stopping = false
	r.state = "stopped"
	r.mu.Unlock()

	return nil
}

// RestartAgent reinicia un agente específico
func (m *Manager) RestartAgent(name string) error {
	m.StopAgent(name)
	return m.StartAgent(name)
}

// AgentStatus representa el estado actual de un agente
type AgentStatus struct {
	Name     string
	State    string
	Restarts int
	LastErr  error
}

// ListAgents devuelve una lista con el estado de todos los agentes
func (m *Manager) ListAgents() []AgentStatus {
	m.mu.Lock()
	var names []string
	for name := range m.agents {
		names = append(names, name)
	}
	m.mu.Unlock()

	var result []AgentStatus
	for _, name := range names {
		m.mu.Lock()
		r, exists := m.agents[name]
		if !exists {
			m.mu.Unlock()
			continue
		}
		m.mu.Unlock()

		r.mu.Lock()
		status := AgentStatus{
			Name:     name,
			State:    r.state,
			Restarts: r.restarts,
			LastErr:  r.lastErr,
		}
		r.mu.Unlock()
		result = append(result, status)
	}
	return result
}

// StartAll inicia todos los agentes registrados
func (m *Manager) StartAll() {
	m.mu.Lock()
	var names []string
	for name := range m.agents {
		names = append(names, name)
	}
	m.mu.Unlock()

	for _, name := range names {
		m.StartAgent(name)
	}
}

// StopAll detiene todos los agentes registrados
func (m *Manager) StopAll() {
	m.mu.Lock()
	var names []string
	for name := range m.agents {
		names = append(names, name)
	}
	m.mu.Unlock()

	for _, name := range names {
		m.StopAgent(name)
	}
}
