package databasepkg

import "time"

type MessageData struct {
	Id        string
	Type      int
	Owner     string
	Text      string
	CreatedAt time.Time
}
