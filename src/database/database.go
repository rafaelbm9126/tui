package databasepkg

import (
	"database/sql"
	"log/slog"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	toolspkg "main/src/tools"
)

type Database struct {
	logger *slog.Logger
	conn   *sql.DB
}

func NewDatabase(logger *slog.Logger) (*Database, error) {
	err := toolspkg.CreateDirIfNotExist(".cache")
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", ".cache/data.db")
	if err != nil {
		return nil, err
	}

	return &Database{logger: logger, conn: db}, nil
}

func (db *Database) Migration() error {
	_, err := db.conn.Exec(`
		PRAGMA journal_mode=WAL;
		PRAGMA synchronous=NORMAL;
		PRAGMA foreign_keys=ON;
		PRAGMA cache_size=-10000;

		CREATE TABLE IF NOT EXISTS messages (
			id text PRIMARY KEY NOT NULL,
			type INTEGER NOT NULL,
			owner TEXT NOT NULL,
			text TEXT NOT NULL,
			created_at text NOT NULL
		)
	`)
	if err != nil {
		db.logger.Error("Error Database [Migration]", "msg", err.Error())
		return err
	}
	return nil
}

func (db *Database) CreateMessage(msg MessageData) (string, error) {
	uid, _ := uuid.NewRandom()

	id := uid.String()

	_, err := db.conn.Exec(`
			INSERT INTO messages (id, type, owner, text, created_at) VALUES (?, ?, ?, ?, ?)
		`,
		id,
		msg.Type,
		msg.Owner,
		msg.Text,
		msg.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		db.logger.Error("Error Database [CreateMessage]", "msg", err.Error())
		return "", err
	}
	return id, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}
