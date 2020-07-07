package gode

import (
	"sync"

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
	clients sync.Map
}

func NewClientHub() *ClientHub {
	return &ClientHub{}
}

func (h *ClientHub) Register(client *client.Client) (err error) {
	h.clients.Store(client, true)

	return
}

func (h *ClientHub) Unregister(client *client.Client) {
	if _, ok := h.clients.Load(client); ok {
		h.clients.Delete(client)
		//log.Printf("client deleted from hub, now have %d clients\n", len(clients.clients))
	}
}

func (h *ClientHub) NumberOfClients() (numbers int) {
	h.clients.Range(func(_, _ interface{}) bool {
		numbers++
		return true
	})

	return
}
