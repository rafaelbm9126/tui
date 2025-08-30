package messagepkg

import (
	"time"

	databasepkg "main/src/database"
	modelpkg "main/src/model"
	toolspkg "main/src/tools"
)

type Database = databasepkg.Database
type MessageModel = modelpkg.MessageModel
type MessageType = modelpkg.MessageType
type MessageSource = modelpkg.MessageSource
type ThreadModel = modelpkg.ThreadModel

type MessageList struct {
	db       *Database
	Messages []MessageModel
	Thread   *ThreadModel
	Threads  []ThreadModel
}

func NewMessageList(db *Database) *MessageList {
	threads, _ := db.ListThreads()

	return &MessageList{
		db:       db,
		Messages: []MessageModel{},
		Thread:   nil,
		Threads:  threads,
	}
}

func NewMessage(
	thread_id string,
	mtype MessageType,
	source MessageSource,
	written_by string,
	text string,
) MessageModel {
	return MessageModel{
		Type:      mtype,
		Source:    source,
		WrittenBy: written_by,
		Text:      text,
		ThreadId:  thread_id,
		CreatedAt: time.Now(),
	}
}

func (ml *MessageList) ControlThread(name string) {
	if ml.Thread != nil {
		return
	}

	ml.Thread, _ = ml.db.CreateThread(ThreadModel{
		Name:      toolspkg.CutString(name, 0, 10),
		CreatedAt: time.Now(),
	})
}

func (ml *MessageList) AddMessage(message MessageModel) {
	message.CreatedAt = time.Now()
	if message.Type != modelpkg.TyCommand && message.Type != modelpkg.TySystem {
		ml.ControlThread(message.Text)
	}
	if ml.Thread != nil {
		message.ThreadId = ml.Thread.Id
		ml.db.CreateMessage(message)
	}
	ml.Messages = append(ml.Messages, message)
}
