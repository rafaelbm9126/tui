package event

import (
	"time"
)

type EventType int

const (
	EvtSystem EventType = iota
	EvtMessage
)

type EventModel struct {
	Type EventType
	Data any
	Time time.Time
}

type Event struct{ evt EventModel }
