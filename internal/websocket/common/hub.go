package common

import (
	"sync"
)

type Hub struct {
	id                string
	onHubHasNoClients func(id string)
	clients           *sync.Map
}

func NewHub(id string, notifyOnEmpty func(id string)) *Hub {
	h := &Hub{
		id:                id,
		clients:           &sync.Map{},
		onHubHasNoClients: notifyOnEmpty,
	}
	return h
}

func (h *Hub) registerClient(c *Client) {
	h.clients.Store(c, struct{}{})
}

func (h *Hub) unregisterClient(c *Client) {
	_, ok := h.clients.LoadAndDelete(c)
	if !ok {
		return
	}
	close(c.send)
	if h.isEmpty() {
		h.onHubHasNoClients(h.id)
	}
}

func (h *Hub) sendToClient(c *Client, msg []byte) {
	_, ok := h.clients.Load(c)
	if !ok {
		return
	}
	select {
	case c.send <- msg:
	default:
		h.unregisterClient(c)
	}
}

func (h *Hub) broadcastMsg(msg []byte) {
	h.clients.Range(func(key, value any) bool {
		c := key.(*Client)
		h.sendToClient(c, msg)
		return true
	})
}

func (h *Hub) isEmpty() bool {
	isEmpty := true
	h.clients.Range(func(key, value any) bool {
		isEmpty = false
		return false
	})
	return isEmpty
}
