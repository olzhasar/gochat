package main

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
