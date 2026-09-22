package websocket

import (
	"encoding/json"
	"log"
	"sync"
)

type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type Message struct {
	Room    string          `json:"room"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Sender  *Client         `json:"-"`
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("Client connected: %s", client.ID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				for room := range client.rooms {
					if clients, ok := h.rooms[room]; ok {
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.rooms, room)
						}
					}
				}
			}
			h.mu.Unlock()
			log.Printf("Client disconnected: %s", client.ID)

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.rooms[message.Room]; ok {
				for client := range clients {
					if message.Sender != nil && client == message.Sender {
						continue
					}
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) BroadcastToRoom(room string, message Message) {
	message.Room = room
	h.broadcast <- message
}

func (h *Hub) JoinRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[*Client]bool)
	}
	h.rooms[room][client] = true
	client.rooms[room] = true
	log.Printf("Client %s joined room %s", client.ID, room)
}

func (h *Hub) LeaveRoom(client *Client, room string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.rooms[room]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}
	delete(client.rooms, room)
	log.Printf("Client %s left room %s", client.ID, room)
}

func (h *Hub) GetRoomCount(room string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.rooms[room]; ok {
		return len(clients)
	}
	return 0
}

func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
