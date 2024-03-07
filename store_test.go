package main

import (
	"testing"
)

func runStoreTests(t *testing.T, store Store) {
	err := store.AddMessage("hello")
	if err != nil {
		t.Fatal(err)
	}

	messages, err := store.GetMessages()
	if err != nil {
		t.Fatal(err)
	}

	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	if messages[0] != "hello" {
		t.Fatalf("expected message 'hello', got %s", messages[0])
	}
}

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()
	runStoreTests(t, store)
}

func TestSQLStore(t *testing.T) {
	store, err := NewSQLStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	runStoreTests(t, store)
}
