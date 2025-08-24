package main

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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

func (ml *MessageList) AddMessage(msg MessageModel) {
	ml.Messages = append(ml.Messages, msg)
}

func (ml *MessageList) AddMessageHuman(text string) {
	ml.Messages = append(ml.Messages, MessageModel{
		Type: Human,
		Text: text,
		Time: time.Now(),
	})
}

func (ml *MessageList) AddMessageAssistant(text string) {
	ml.Messages = append(ml.Messages, MessageModel{
		Type: Assistant,
		Text: text,
		Time: time.Now(),
	})
}

func (ml *MessageList) PrePrintMessages(t *TUI) string {
	var sb strings.Builder

	for _, message := range ml.Messages {
		sb.WriteString(t.DottedLine(t.width) + "\n")

		// Message Header //
		var label lipgloss.Style
		header := message.Type.String()
		if !message.Time.IsZero() {
			header += " - " + message.Time.Format("15:04:05")
		}
		switch message.Type {
		case System:
			label = t.styles.labelSystem
		case Human:
			label = t.styles.labelHuman
		case Assistant:
			label = t.styles.labelAssistant
		}
		sb.WriteString(label.Render(header) + "\n")
		// [End] Message Header //

		// Message Body //
		body := message.Text
		if t.mdEnabled && t.mdRenderer != nil {
			rendered, err := t.mdRenderer.Render(body)
			if err == nil {
				body = strings.TrimSpace(rendered)
			}
		}
		sb.WriteString(t.styles.body.Render(body) + "\n")
		// [End] Message Body //
	}

	return sb.String()
}
