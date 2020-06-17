package gode

import (
	"fmt"

	"gode/client"
)

const MaxClients = 100

type Hub struct {
	// Registered clients.
	clients map[*client.Client]bool
}

func NewHub() *Hub {
	return &Hub{
		//register: make(chan *Client),
		clients: make(map[*client.Client]bool),
	}
}

func (h *Hub) register(client *client.Client) (err error) {
	if len(h.clients) < MaxClients {
		h.clients[client] = true
	} else {
		return fmt.Errorf("client full")
	}
	//log.Printf("new client add to hub, now have %d clients\n", len(h.clients))

	return
}

func (h *Hub) unregister(client *client.Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		//log.Printf("client deleted from hub, now have %d clients\n", len(h.clients))
	}
}

func (h *Hub) NumberOfClients() int {
	return len(h.clients)
}
