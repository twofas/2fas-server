package tests

import (
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"log"
	"net/url"
	"testing"
)

type WebsocketTestListener struct {
	ListenAddr       *url.URL
	ReceivedMessages chan string
}

func NewWebsocketTestListener(uri string) *WebsocketTestListener {
	addr, _ := url.Parse("ws://localhost:8081/" + uri)

	receivedMessages := make(chan string)

	return &WebsocketTestListener{
		ListenAddr:       addr,
		ReceivedMessages: receivedMessages,
	}
}

func (l *WebsocketTestListener) StartListening() *websocket.Conn {
	c, _, err := websocket.DefaultDialer.Dial(l.ListenAddr.String(), nil)

	if err != nil {
		log.Fatal("dial:", err)
	}

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()

			if err != nil {
				log.Println("read:", err)
				return
			}

			l.ReceivedMessages <- string(message)
		}
	}()

	return c
}

func (l *WebsocketTestListener) AssertMessageHasBeenReceived(t *testing.T, expected string) {
	assert.JSONEq(t, expected, <-l.ReceivedMessages)
}
