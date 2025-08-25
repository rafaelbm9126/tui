package messagepkg

import (
	"time"

	databasepkg "main/src/database"
)

type MessageType int

const (
	System MessageType = iota
	Human
	Assistant
)

func (d MessageType) String() string {
	return [...]string{
		"System",
		"Human",
		"Assistant",
	}[d]
}

type MessageModel struct {
	Type MessageType
	From string
	Text string
	Time time.Time
}

type MessageList struct {
	db       *databasepkg.Database
	Messages []MessageModel
}

func NewMessageList(db *databasepkg.Database) *MessageList {
	return &MessageList{
		db:       db,
		Messages: []MessageModel{},
	}
}

func (ml *MessageList) AddMessageSystem(text string) {
	ml.Messages = append(ml.Messages, MessageModel{
		Type: System,
		Text: text,
		Time: time.Now(),
	})
}

func (ml *MessageList) AddMessageHuman(text string) {
	msg := MessageModel{
		Type: Human,
		Text: text,
		From: "Human",
		Time: time.Now(),
	}

	ml.Messages = append(ml.Messages, msg)

	ml.db.CreateMessage(databasepkg.MessageData{
		Type:      int(msg.Type),
		Owner:     msg.From,
		Text:      msg.Text,
		CreatedAt: msg.Time,
	})
}

func (ml *MessageList) AddMessageAssistant(text string, from string) {
	msg := MessageModel{
		Type: Assistant,
		From: from,
		Text: text,
		Time: time.Now(),
	}

	ml.Messages = append(ml.Messages, msg)

	ml.db.CreateMessage(databasepkg.MessageData{
		Type:      int(msg.Type),
		Owner:     msg.From,
		Text:      msg.Text,
		CreatedAt: msg.Time,
	})
}
