package common

import (
	"fmt"
	"sync"
	"testing"

	"github.com/gorilla/websocket"
)

func TestRegisterClientDoesNotCreateSameHubTwice(t *testing.T) {
	hp := newHubPool()

	const channelID = "channelID"

	_, h1 := hp.registerClient(channelID, &websocket.Conn{})
	_, h2 := hp.registerClient(channelID, &websocket.Conn{})

	if h1 != h2 {
		t.Fatal("New hub was created")
	}
}

func TestRemovingEmptyHub(t *testing.T) {
	hp := newHubPool()

	const channelID = "channelID"
	c, h1 := hp.registerClient(channelID, &websocket.Conn{})
	h1.unregisterClient(c)

	_, h2 := hp.registerClient(channelID, &websocket.Conn{})

	if !h1.isEmpty() {
		t.Fatalf("Hub does not report it is empty, even though uit should")
	}
	if h1 == h2 {
		t.Fatal("Old hub wasn't deleted")
	}
	if h2.isEmpty() {
		t.Fatal("New heb is empty, even though it shouldn't")
	}
}

// TestCreateRemoveConcurrently in which we (for each channel) register a client and then unregister it immediately.
// The last client to be register stays that way.
// We then check:
// - if poll has non-empty hubs,
// - iff all hubs removed from the poll are empty.
func TestCreateRemoveConcurrently(t *testing.T) {
	hp := newHubPool()
	const channelsNo = 100
	const clientsPerChannel = 1000

	hubs := &sync.Map{}

	wg := sync.WaitGroup{}
	wg.Add(channelsNo * clientsPerChannel)
	for i := 0; i < channelsNo; i++ {
		var channelID = fmt.Sprintf("channel-%d", i)
		go func() {
			for j := 0; j < clientsPerChannel; j++ {
				c, h := hp.registerClient(channelID, &websocket.Conn{})
				hubs.Store(h, struct{}{})
				go func() {
					h.unregisterClient(c)
					wg.Done()
				}()
			}
			_, h := hp.registerClient(channelID, &websocket.Conn{})
			hubs.Store(h, struct{}{})
		}()
	}

	wg.Wait()

	for c, hub := range hp.hubs {
		if hub.isEmpty() {
			t.Fatalf("Empty hub found in channel: %q", c)
		}
	}

	hubs.Range(func(key, value any) bool {
		h1 := key.(*Hub)
		if !h1.isEmpty() {
			if h2, ok := hp.hubs[h1.id]; !ok || h1 != h2 {
				t.Fatalf("Non-empty hub was evicted from hub pool: %q", h1.id)
			}
		}
		return true
	})
}
