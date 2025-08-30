package databasepkg

import (
	"database/sql"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"

	modelpkg "main/src/model"
	toolspkg "main/src/tools"
)

type MessageModel = modelpkg.MessageModel
type ThreadModel = modelpkg.ThreadModel

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

		CREATE TABLE IF NOT EXISTS threads (
			id TEXT PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY NOT NULL,
			type INTEGER NOT NULL,
			source INTEGER NOT NULL,
			written_by TEXT NOT NULL,
			text TEXT NOT NULL,
			thread_id TEXT NOT NULL,
			created_at TEXT NOT NULL,

			FOREIGN KEY (thread_id) REFERENCES threads(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		db.logger.Error("Error Database [Migration]", "msg", err.Error())
		return err
	}
	return nil
}

func (db *Database) CreateThread(thd ThreadModel) (*ThreadModel, error) {
	thd.Id = toolspkg.GenerateUUID()

	_, err := db.conn.Exec(`
			INSERT INTO threads (id, name, created_at) VALUES (?, ?, ?)
		`,
		thd.Id,
		thd.Name,
		thd.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		db.logger.Error("Error Database [CreateThread]", "msg", err.Error())
		return nil, err
	}

	return &thd, nil
}

func (db *Database) ListThreads() ([]ThreadModel, error) {
	rows, err := db.conn.Query(`
			SELECT id, name, created_at FROM threads ORDER BY created_at ASC
		`)
	if err != nil {
		db.logger.Error("Error Database [ListThreads]", "msg", err.Error())
		return nil, err
	}
	defer rows.Close()

	var threads []ThreadModel
	for rows.Next() {
		var thd ThreadModel
		var createdAt string

		if err := rows.Scan(&thd.Id, &thd.Name, &createdAt); err != nil {
			db.logger.Error("Error Database [ListThreads]", "msg", err.Error())
			return nil, err
		}

		parsedTime, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			db.logger.Error("Error Database [ListThreads] parsing time", "msg", err.Error())
			return nil, err
		}
		thd.CreatedAt = parsedTime

		threads = append(threads, thd)
	}
	if err := rows.Err(); err != nil {
		db.logger.Error("Error Database [ListThreads]", "msg", err.Error())
		return nil, err
	}

	return threads, nil
}

func (db *Database) UpdateThread(thd ThreadModel) error {
	_, err := db.conn.Exec(`
			UPDATE threads SET name = ? WHERE id = ?
		`,
		thd.Name,
		thd.Id,
	)
	if err != nil {
		db.logger.Error("Error Database [UpdateThread]", "msg", err.Error())
		return err
	}

	return nil
}

func (db *Database) DeleteThread(thd ThreadModel) error {
	_, err := db.conn.Exec(`
			DELETE FROM threads WHERE id = ?
		`,
		thd.Id,
	)
	if err != nil {
		db.logger.Error("Error Database [DeleteThread]", "msg", err.Error())
		return err
	}

	return nil
}

func (db *Database) CreateMessage(msg MessageModel) (string, error) {
	id := toolspkg.GenerateUUID()

	_, err := db.conn.Exec(`
			INSERT INTO messages (
				id,
				type,
				source,
				written_by,
				text,
				thread_id,
				created_at
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`,
		id,
		msg.Type,
		msg.Source,
		msg.WrittenBy,
		msg.Text,
		msg.ThreadId,
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
