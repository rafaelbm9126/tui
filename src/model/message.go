package modelpkg

import (
	"time"
)

/**
 * MESSAGE MODEL
 */

type MessageType int

const (
	TySystem MessageType = iota
	TyText
	TyCommand
)

func (mt MessageType) String() string {
	return [...]string{
		"System",
		"Text",
		"Command",
	}[mt]
}

type MessageSource int

const (
	ScSystem MessageSource = iota
	ScHuman
	ScAssistant
)

func (mf MessageSource) String() string {
	return [...]string{
		"System",
		"Human",
		"Assistant",
	}[mf]
}

type MessageModel struct {
	Type      MessageType
	Source    MessageSource
	WrittenBy string
	Text      string
	ThreadId  string
	CreatedAt time.Time
}

/**
 * THREAD MODEL
 */

type ThreadModel struct {
	Id        string
	Name      string
	CreatedAt time.Time
}
