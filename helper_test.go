package gode_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gode/client"
)

type SpyCaller struct {
	history  apiHistory
	response map[string][]byte
}

func (c *SpyCaller) Call(service string, functionName string, parameters ...interface{}) ([]byte, error) {
	c.history = append(c.history, apiLog{
		service:    service,
		function:   functionName,
		parameters: parameters,
	})

	return c.response[functionName], nil
}

type apiHistory []apiLog

type apiLog struct {
	service    string
	function   string
	parameters []interface{}
}

type SpyHub struct {
	clients []*client.Client
}

func (h *SpyHub) NumberOfClients() int {
	return len(h.clients)
}

func (h *SpyHub) Register(c *client.Client) error {
	h.clients = append(h.clients, c)

	return nil
}

func (h *SpyHub) Unregister(client *client.Client) {}

func assertReceiveBinaryMsg(t *testing.T, dialer *websocket.Conn, want string) {
	t.Helper()

	mt, p, err := dialer.ReadMessage()
	if err != nil {
		t.Fatal("ReadMessageError", err)
	}
	if mt != websocket.BinaryMessage {
		t.Errorf("expect got message type %d, got %d", websocket.BinaryMessage, mt)
	}
	got := string(p)
	if got != want {
		t.Errorf("message from web socket not matched\nwant %s\n got %s", want, got)
	}
}

func writeBinaryMsg(t *testing.T, wsClient *websocket.Conn, msg string) {
	err := wsClient.WriteMessage(websocket.BinaryMessage, []byte(msg))
	if err != nil {
		t.Error("ws WriteMessage Error", err)
	}
}

func assertWithin(t *testing.T, d time.Duration, assert func()) {
	t.Helper()

	done := make(chan struct{}, 1)

	go func() {
		assert()
		done <- struct{}{}
	}()

	select {
	case <-time.After(d):
		t.Error("timed out")
	case <-done:
	}
}

func assertNoResponseWithin(t *testing.T, d time.Duration, client *websocket.Conn) {
	msgChan := make(chan []byte, 1)
	go func() {
		_, p, _ := client.ReadMessage()
		msgChan <- p
	}()

	select {
	case <-time.After(d):
		return
	case msg := <-msgChan:
		t.Errorf("shouldn't get response but got %q", msg)
	}
}

func assertResponseCode(t *testing.T, got, expected int) {
	t.Helper()
	if got != expected {
		t.Errorf("expect response status code %d, got %d", expected, got)
	}
}

func mustDialWS(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	dialer, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("could not open a ws connection on %s %v", url, err)
	}
	return dialer
}

func makeWebSocketURL(server *httptest.Server, path string) string {
	url := "ws" + strings.TrimPrefix(server.URL, "http") + path
	return url
}

func waitForProcess() {
	time.Sleep(1 * time.Millisecond)
}

func assertNumberOfClient(t *testing.T, wanted, got int) {
	t.Helper()
	if got != wanted {
		t.Errorf("wanted number of clients %d, got, %d", wanted, got)
	}
}
