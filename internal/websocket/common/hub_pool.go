package common

import (
	"sync"

	"github.com/gorilla/websocket"
)

// hubPool manages the creation of hubs and their removal from the pool when they are empty.
//
// All access to the `hubs' map, or to any `hub' in that map, should be done only after obtaining `mtx'.
// Registering a client with a hub can only happen when `mtx` is held. This makes it safe to delete an empty hub.
// Even if some other goroutine runs hub.deregisterClient, there will be nothing to remove.
type hubPool struct {
	hubs map[string]*Hub
	mtx  *sync.Mutex
}

func newHubPool() *hubPool {
	return &hubPool{
		hubs: map[string]*Hub{},
		mtx:  &sync.Mutex{},
	}
}

// registerClient is called by handler.
func (h *hubPool) registerClient(channel string, conn *websocket.Conn) (*Client, *Hub) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	hub := h.getOrCreateHub(channel)
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), sendMtx: &sync.Mutex{}}
	hub.registerClient(client)

	// handler (caller of this method) isn't really interested in hub,
	// but it's useful for testing.
	return client, hub
}

func (h *hubPool) getOrCreateHub(channel string) *Hub {
	hub, ok := h.hubs[channel]
	if !ok {
		hub = NewHub(channel, h.onHubIsHasNoClients)
		h.hubs[channel] = hub
	}

	return hub
}

// onHubIsHasNoClients is called by the hub after if unregistered a client, if it has no clients left.
func (h *hubPool) onHubIsHasNoClients(channel string) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	hub, ok := h.hubs[channel]
	if !ok {
		// Hub was already deleted.
		return
	}
	if !hub.isEmpty() {
		// Between this function was invoked (and mutex acquired), new client was registered.
		// We must skip the deletion.
		return
	}

	delete(h.hubs, channel)
}
