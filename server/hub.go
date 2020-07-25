package main

import (
	"fmt"
)

type Hub struct {
	clients    map[string]*Client
	message    chan Message
	register   chan *Client
	unregister chan *Client
}

type Message struct {
	To      string `json:"To"`
	Message string `json:"Message"`
	From    string `json:"From"`
}

func newHub() *Hub {
	return &Hub{
		message:    make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.userId] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.userId]; ok {
				delete(h.clients, client.userId)
				close(client.send)
			}
		case newMsg := <-h.message:
			targetClient, ok := h.clients[newMsg.To]
			if !ok {
				fmt.Println("client not found")
			}
			select {
			case targetClient.send <- []byte(fmt.Sprintf("%v", newMsg)):
			default:
				close(targetClient.send)
				delete(h.clients, targetClient.userId)
			}
		}
	}
}
