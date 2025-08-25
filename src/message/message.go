package messagepkg

import (
	"time"
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
	Messages []MessageModel
}

func NewMessageList() *MessageList {
	return &MessageList{
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
	ml.Messages = append(ml.Messages, MessageModel{
		Type: Human,
		Text: text,
		Time: time.Now(),
	})
}

func (ml *MessageList) AddMessageAssistant(text string, from string) {
	ml.Messages = append(ml.Messages, MessageModel{
		Type: Assistant,
		From: from,
		Text: text,
		Time: time.Now(),
	})
}
