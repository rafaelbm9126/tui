package main

import (
	"errors"
	"log/slog"
	"sync"
	"time"
)

// Bus define la interfaz para el sistema de comunicación pub/sub
type Bus interface {
	Publish(evtype EventType, data any)
	Subscribe(evtype EventType, buf int) (<-chan Event, func(), error)
	Length(evtype EventType) int
	Close()
}

// MemoryBus implementa Bus en memoria
type OptimizedBus struct {
	mu     sync.RWMutex
	closed bool
	nextID int
	subs   map[EventType]map[int]*sub
	logger *slog.Logger
}

type sub struct {
	id int
	ch chan EventModel
	// out chan Event removed - direct channel
	stop chan struct{}
	// WaitGroup removed - no intermediate goroutine
}

// NewMemoryBus crea una nueva instancia de MemoryBus
func NewMemoryBus(logger *slog.Logger) *OptimizedBus {
	return &OptimizedBus{
		subs:   make(map[EventType]map[int]*sub),
		logger: logger,
	}
}

// Publish publica un evento en el bus
func (b *OptimizedBus) Publish(evtype EventType, data any) {
	evt := EventModel{Type: evtype, Data: data, Time: time.Now()}

	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return
	}
	var targets []*sub
	if m, ok := b.subs[evtype]; ok {
		for _, s := range m {
			targets = append(targets, s)
		}
	}
	b.mu.RUnlock()

	b.logger.Info("Received event >> ", "len", data)

	for _, s := range targets {
		select {
		case s.ch <- evt:
		default:
			select { // drop-oldest
			case <-s.ch:
			default:
			}
			select {
			case s.ch <- evt:
			default:
			}
		}
	}
}

// Subscribe suscribe a un evento en el bus
func (b *OptimizedBus) Subscribe(evtype EventType, buf int) (<-chan EventModel, func(), error) {
	if buf <= 0 {
		buf = 64
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil, nil, errors.New("bus closed")
	}

	// Crea una nueva suscripción
	id := b.nextID
	b.nextID++
	s := &sub{
		id: id,
		ch: make(chan EventModel, buf),
		// Removed intermediate channel
		stop: make(chan struct{}),
	}

	// La añade a la lista de subs por topic
	if b.subs[evtype] == nil {
		b.subs[evtype] = make(map[int]*sub)
	}
	b.subs[evtype][id] = s

	// Función de cancelación (cleanup)
	unsub := func() {
		b.mu.Lock()
		if subs, exists := b.subs[evtype]; exists {
			delete(subs, id)
			if len(subs) == 0 {
				delete(b.subs, evtype)
			}
		}
		b.mu.Unlock()
		close(s.stop) // Cierra el canal de parada para que la goroutine termine
		// No WaitGroup needed with direct channel
	}

	return s.ch, unsub, nil
}

func (b *OptimizedBus) Length(evtype EventType) int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if m, ok := b.subs[evtype]; ok {
		return len(m)
	}
	return 0
}

// Close cierra el bus y todas sus suscripciones
func (b *OptimizedBus) Close() {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return
	}
	b.closed = true

	// Copia todos los subs para poder iterar sin el lock
	var allSubs []*sub
	for _, m := range b.subs {
		for _, s := range m {
			allSubs = append(allSubs, s)
		}
	}
	b.subs = nil
	b.mu.Unlock()

	// Cierra los canales de stop para que las goroutines terminen
	for _, s := range allSubs {
		close(s.stop)
	}
}
