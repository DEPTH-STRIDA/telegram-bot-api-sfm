// Package cache provides implementations of LastMessageCache interface
// for auto-deletion feature.
//
// To use PostgreSQL implementation, install the dependency:
//
//	go get github.com/lib/pq
package cache

import (
	"database/sql"
	"time"

	tgfsm "tgfsm"

	_ "github.com/lib/pq"
)

// PostgresLastMessage implements LastMessageCache using PostgreSQL
type PostgresLastMessage struct {
	db        *sql.DB
	tableName string
}

// Ensure PostgresLastMessage implements LastMessageCache interface
var _ tgfsm.LastMessageCache = (*PostgresLastMessage)(nil)

// NewPostgresLastMessage creates a new cache implementation using PostgreSQL
// db - database connection
// tableName - table name (default: "last_messages")
func NewPostgresLastMessage(db *sql.DB, tableName string) (*PostgresLastMessage, error) {
	if tableName == "" {
		tableName = "last_messages"
	}

	cache := &PostgresLastMessage{
		db:        db,
		tableName: tableName,
	}

	// Create table if it doesn't exist
	if err := cache.createTable(); err != nil {
		return nil, err
	}

	return cache, nil
}

// createTable creates table for storing last messages
func (c *PostgresLastMessage) createTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS ` + c.tableName + ` (
		chat_id BIGINT PRIMARY KEY,
		message_id INTEGER NOT NULL,
		important BOOLEAN NOT NULL DEFAULT FALSE,
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);
	
	CREATE INDEX IF NOT EXISTS idx_` + c.tableName + `_updated_at ON ` + c.tableName + `(updated_at);
	`
	_, err := c.db.Exec(query)
	return err
}

// GetLastMessageInfo returns information about the last message for a chat
func (c *PostgresLastMessage) GetLastMessageInfo(chatID int64) (*tgfsm.LastMessageInfo, error) {
	query := "SELECT message_id, important FROM " + c.tableName + " WHERE chat_id = $1"
	var messageID int
	var important bool
	err := c.db.QueryRow(query, chatID).Scan(&messageID, &important)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tgfsm.LastMessageInfo{
		MessageID: messageID,
		Important: important,
	}, nil
}

// SetLastMessageInfo saves information about the last message for a chat
func (c *PostgresLastMessage) SetLastMessageInfo(chatID int64, info *tgfsm.LastMessageInfo) error {
	query := `
	INSERT INTO ` + c.tableName + ` (chat_id, message_id, important, updated_at)
	VALUES ($1, $2, $3, NOW())
	ON CONFLICT (chat_id) 
	DO UPDATE SET message_id = $2, important = $3, updated_at = NOW()
	`
	_, err := c.db.Exec(query, chatID, info.MessageID, info.Important)
	return err
}

// DeleteLastMessageInfo deletes information about the last message for a chat
func (c *PostgresLastMessage) DeleteLastMessageInfo(chatID int64) error {
	query := "DELETE FROM " + c.tableName + " WHERE chat_id = $1"
	_, err := c.db.Exec(query, chatID)
	return err
}

// CleanupOldRecords deletes records older than specified time
// Useful for periodic cleanup of old data
func (c *PostgresLastMessage) CleanupOldRecords(olderThan time.Duration) error {
	query := "DELETE FROM " + c.tableName + " WHERE updated_at < NOW() - $1::INTERVAL"
	_, err := c.db.Exec(query, olderThan)
	return err
}
