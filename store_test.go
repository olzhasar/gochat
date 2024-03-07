package main

import (
	"testing"
)

func TestAddMessage(t *testing.T) {
	store := NewMemoryStore()
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
