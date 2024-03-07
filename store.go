package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Store interface {
	AddMessage(message string) error
	GetMessages() ([]string, error)
}

type MemoryStore struct {
	messages []string
}

func (m *MemoryStore) AddMessage(message string) error {
	m.messages = append(m.messages, message)
	return nil
}

func (m *MemoryStore) GetMessages() ([]string, error) {
	return m.messages, nil
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

type SQLStore struct {
	db *sql.DB
}

func (s *SQLStore) AddMessage(message string) error {
	_, err := s.db.Exec("INSERT INTO messages (message) VALUES (?)", message)
	return err
}

func (s *SQLStore) GetMessages() ([]string, error) {
	rows, err := s.db.Query("SELECT message FROM messages")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []string
	for rows.Next() {
		var message string
		if err := rows.Scan(&message); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func NewSQLStore(dbPath string) (*SQLStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS messages (message TEXT)")
	if err != nil {
		return nil, err
	}

	return &SQLStore{db: db}, nil
}
