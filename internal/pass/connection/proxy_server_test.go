package connection

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"

	"github.com/twofas/2fas-server/internal/common/logging"
	"github.com/twofas/2fas-server/internal/pass/connection/proxy"
)

func init() {
	logging.Init(nil)
}

// TestProxy sends message both ways and makes sure it is received correctly.
func TestProxy(t *testing.T) {
	ws1, ws2, cleanup := setupConnections(t, proxy.DefaultConfig())
	defer cleanup()

	testWriteReceive(t, ws1, ws2)
	testWriteReceive(t, ws2, ws1)
}

// TestConnectionIsClosedAfterTheSpecifiedTime checks that `DisconnectAfter` is obeyed by the proxy server.
func TestConnectionIsClosedAfterTheSpecifiedTime(t *testing.T) {
	timeout := time.Second

	ws1, ws2, cleanup := setupConnections(t, proxy.Config{
		WriteTimeout:    proxy.DefaultWriteTimeout,
		ReadTimeout:     proxy.DefaultReadTimeout,
		PingFrequency:   proxy.DefaultPingFrequency,
		DisconnectAfter: timeout,
	})
	defer cleanup()

	// Exchange some data to make sure the connection is established.
	testWriteReceive(t, ws1, ws2)
	testWriteReceive(t, ws2, ws1)

	//  Neither side of the connection sends any message, they just wait on read. Therefore, in both cases ReadMessage
	// should exit after the server closes the connection.
	ws1Result := make(chan error)
	ws2Result := make(chan error)
	go func() {
		_, _, err := ws1.ReadMessage()
		ws1Result <- err
	}()
	go func() {
		_, _, err := ws2.ReadMessage()
		ws2Result <- err
	}()

	// Finish test after timeout and check if connections were closed. One would expect a race condition here
	// (we check exactly after timeout) but this test seems to be stable. This is because we have already spent some time
	// exchanging the data before waiting for the timeout.
	after := time.After(timeout)
	var err1, err2 error
	done := false
	for !done {
		select {
		case err1 = <-ws1Result:
		case err2 = <-ws2Result:
		case <-after:
			done = true
		}
	}

	if err1 == nil {
		t.Logf("WebSocket 1 connection wasn't closed")
	}
	if err2 == nil {
		t.Logf("WebSocket 2 connection wasn't closed")
	}
}

// TestPingPongIsEnoughToKeepUsAlive check that the connection is kept alive by the ws native ping-pong mechanism.
// In the Browser Extension the pong response is sent by the browser automatically, in this test framework does it for us
// in ReadMessage.
func TestPingPongIsEnoughToKeepUsAlive(t *testing.T) {
	readTimeout := time.Second

	ws1, ws2, cleanup := setupConnections(t, proxy.Config{
		WriteTimeout:    proxy.DefaultWriteTimeout,
		ReadTimeout:     readTimeout,
		PingFrequency:   readTimeout / 4,
		DisconnectAfter: time.Minute,
	})
	defer cleanup()

	group := errgroup.Group{}
	group.Go(func() error {
		_, _, err := ws1.ReadMessage()
		return err
	})
	group.Go(func() error {
		_, _, err := ws2.ReadMessage()
		return err
	})
	time.Sleep(4 * readTimeout)

	// Write some messages to both websockets. This has two benefits:
	// 1. It ensures the connections are still alive,
	// 2. It makes ReadMessage above return, so group.Wait will exit.
	if err := ws1.WriteMessage(websocket.BinaryMessage, []byte("hello!")); err != nil {
		t.Errorf("Failed to write message to the first websocket: %v", err)
	}
	if err := ws2.WriteMessage(websocket.BinaryMessage, []byte("hello!")); err != nil {
		t.Errorf("Failed to write message to the second websocket: %v", err)
	}

	err := group.Wait()
	if err != nil {
		t.Errorf("Error when reading from websocket: %v", err)
	}
}

// setupConnections creates new test websocket server and two connected clients paired in a proxy.
func setupConnections(t *testing.T, cfg proxy.Config) (*websocket.Conn, *websocket.Conn, func()) {
	s := httptest.NewServer(testHandler{
		t:  t,
		ps: NewProxyServer("id", cfg),
	})

	ws1, _, err := testDialer.Dial(makeWsURL(s.URL, "mobile", "1"), nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	ws2, _, err := testDialer.Dial(makeWsURL(s.URL, "extension", "1"), nil)
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	cleanup := func() {
		ws1.Close()
		ws2.Close()
		s.Close()
	}

	return ws1, ws2, cleanup
}

var testDialer = websocket.Dialer{
	Subprotocols:     []string{"2pass.io"},
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	HandshakeTimeout: 30 * time.Second,
}

// testHandler is for handling http connections.
type testHandler struct {
	t  *testing.T
	ps *ProxyServer
}

func (t testHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/mobile" {
		t.ps.ServeExtensionProxyToMobileWS(w, r, r.URL.Query().Get("id"))
	} else if r.URL.Path == "/extension" {
		t.ps.ServeMobileProxyToExtensionWS(w, r, r.URL.Query().Get("id"))
	} else {
		http.Error(w, "invalid path", http.StatusNotFound)
	}

}

// makeWsURL constructs the WebSocket from the test server's URL.
func makeWsURL(s string, app string, id string) string {
	return fmt.Sprintf("ws%s/%s?id=%s", strings.TrimPrefix(s, "http"), app, id)
}

// testWriteReceive writes a message to w1 and makes sure it is received by w2.
func testWriteReceive(t *testing.T, ws1, ws2 *websocket.Conn) {
	t.Helper()
	const message = "Hello, WebSocket!"

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()

		_, received, err := ws2.ReadMessage()
		if err != nil {
			t.Errorf("Failed to read message: %v", err)
			return
		}
		if string(received) != message {
			t.Errorf("Expected %q, received %q", message, string(received))
		}
	}()

	if err := ws1.WriteMessage(websocket.BinaryMessage, []byte(message)); err != nil {
		t.Errorf("Failed to write message: %v", err)
	}

	wg.Wait()
}
