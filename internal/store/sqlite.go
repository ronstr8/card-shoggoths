package store

import (
	"card-shoggoths/internal/game"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Create table
	query := `
	CREATE TABLE IF NOT EXISTS games (
		id TEXT PRIMARY KEY,
		state TEXT,
		updated_at DATETIME
	);
	`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) Save(id string, state *game.GameState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO games (id, state, updated_at) VALUES (?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET state=excluded.state, updated_at=excluded.updated_at;
	`
	_, err = s.db.Exec(query, id, string(data), time.Now())
	return err
}

func (s *SQLiteStore) Load(id string) (*game.GameState, error) {
	var data string
	err := s.db.QueryRow("SELECT state FROM games WHERE id = ?", id).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}

	var state game.GameState
	if err := json.Unmarshal([]byte(data), &state); err != nil {
		return nil, err
	}
	return &state, nil
}
