package message

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"main/src/command"
	"main/src/tui"
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
	command  *Command
}

func NewMessageList(command *Command) *MessageList {
	return &MessageList{
		Messages: []MessageModel{},
		command:  command,
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

	ml.command.IsCommandThenRun(text)
}

func (ml *MessageList) AddMessageAssistant(text string, from string) {
	ml.Messages = append(ml.Messages, MessageModel{
		Type: Assistant,
		From: from,
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
		switch message.Type {
		case System:
			label = t.styles.labelSystem
		case Human:
			label = t.styles.labelHuman
		case Assistant:
			label = t.styles.labelAssistant
			header += " [" + message.From + "]"
		}
		if !message.Time.IsZero() {
			header += " - " + message.Time.Format("15:04:05")
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
