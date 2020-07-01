package gode

import (
	"fmt"

	"gode/client"
)

const MaxClients = 100

type ClientPool interface {
	NumberOfClients() int
	Register(*client.Client) error
	Unregister(*client.Client)
}

type ClientHub struct {
	// Registered clients.
	clients map[*client.Client]bool
}

func NewClientHub() *ClientHub {
	return &ClientHub{
		//register: make(chan *Client),
		clients: make(map[*client.Client]bool),
	}
}

func (h *ClientHub) Register(client *client.Client) (err error) {
	if len(h.clients) < MaxClients {
		h.clients[client] = true
	} else {
		return fmt.Errorf("client full")
	}
	//log.Printf("new client add to hub, now have %d clients\n", len(clients.clients))

	return
}

func (h *ClientHub) Unregister(client *client.Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		//log.Printf("client deleted from hub, now have %d clients\n", len(clients.clients))
	}
}

func (h *ClientHub) NumberOfClients() int {
	return len(h.clients)
}
