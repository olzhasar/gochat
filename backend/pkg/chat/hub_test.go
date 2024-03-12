package chat

import "testing"

func TestHubCreateRoom(t *testing.T) {
	hub := NewHub()

	room1 := hub.CreateRoom()

	if room1 == nil {
		t.Fatal("expected room to be created")
	}

	if room1.ID == "" {
		t.Fatal("expected room to have an ID")
	}

	if hub.GetRoom(room1.ID) != room1 {
		t.Fatal("expected room to be retrievable")
	}

	room2 := hub.CreateRoom()

	if room2.ID == room1.ID {
		t.Fatal("expected rooms to have unique IDs")
	}

	if hub.RoomCount() != 2 {
		t.Fatal("expected hub to have 2 rooms")
	}
}

func TestHubRegisterClient(t *testing.T) {
	hub := NewHub()
	hub.Run()

	room := hub.CreateRoom()

	client := NewClient(nil)
	hub.Register(client, room)

	if len(room.clients) != 1 {
		t.Fatal("expected client to be in room")
	}

	if room.clients[0] != client {
		t.Fatal("expected client to be in room")
	}
}
