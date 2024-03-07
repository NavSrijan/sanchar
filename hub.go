package main

import (
	"fmt"
	"gorm.io/gorm"
)

type Hub struct {
	clients    map[int]map[*Client]bool
	unregister chan *Client
	register   chan *Client
	broadcast  chan Message
	db		   *gorm.DB 
}

type Message struct {
	Type      string `json:"type"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
	ID        int	 `json:"id"`
}

func NewHub(db *gorm.DB) *Hub {
	return &Hub{
		clients:    make(map[int]map[*Client]bool),
		unregister: make(chan *Client),
		register:   make(chan *Client),
		broadcast:  make(chan Message),
		db:         db,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.RegisterNewClient(client)
		case client := <-h.unregister:
			h.RemoveClient(client)
		case message := <-h.broadcast:
			h.HandleMessage(message)

		}
	}
}

func (h *Hub) RegisterNewClient(client *Client) {
	connections := h.clients[client.ID]
	if connections == nil {
		connections = make(map[*Client]bool)
		h.clients[client.ID] = connections
	}
	h.clients[client.ID][client] = true

	fmt.Println("Size of clients: ", len(h.clients[client.ID]))
}

func (h *Hub) RemoveClient(client *Client) {
	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients[client.ID], client)
		close(client.send)
		fmt.Println("Removed client")
	}
}

func (h *Hub) HandleMessage(message Message) {
	fmt.Println("Message received: ", message)
	if message.Type == "message" {
		sender_id, err := get_user_by_username(h.db, message.Sender)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}

		receiver_id, err := get_user_by_username(h.db, message.Recipient)
		message.ID = get_convo(h.db, uint(sender_id.User_id), uint(receiver_id.User_id))
		fmt.Println("Convo ID: ", message.ID)
		clients := h.clients[message.ID]
		for client := range clients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(h.clients[message.ID], client)
			}
		}
	}

}


