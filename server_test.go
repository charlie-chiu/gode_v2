package gode_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gode"
)

func TestConnectionPool(t *testing.T) {
	hub := gode.NewHub()
	Server := httptest.NewServer(gode.NewServer(hub))

	t.Run("NumberOfClients return number of clients from hub", func(t *testing.T) {
		assertNumberOfClient(t, 0, hub.NumberOfClients())
	})

	_ = mustDialWS(t, makeWebSocketURL(Server, "/casino/5145"))
	_ = mustDialWS(t, makeWebSocketURL(Server, "/casino/5145"))
	_ = mustDialWS(t, makeWebSocketURL(Server, "/casino/5145"))
	defer Server.Close()

	t.Run("register when client connect", func(t *testing.T) {
		assertNumberOfClient(t, 3, hub.NumberOfClients())
	})
}

func assertNumberOfClient(t *testing.T, wanted, got int) {
	t.Helper()
	if got != wanted {
		t.Errorf("wanted number of clients %d, got, %d", wanted, got)
	}
}

func TestHTTPRequest(t *testing.T) {
	t.Run("/ returns 404", func(t *testing.T) {
		server := gode.NewServer(gode.NewHub())

		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, request)

		assertResponseCode(t, recorder.Code, http.StatusNotFound)
	})

	t.Run("get /casino/5145 returns 400 bad request", func(t *testing.T) {
		server := gode.NewServer(gode.NewHub())

		request, _ := http.NewRequest(http.MethodGet, "/casino/5145", nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusBadRequest)
	})
}

func TestWebSocket(t *testing.T) {
	const timeout = time.Second

	t.Run("/casino/5145 send ws msg 5145 on connect", func(t *testing.T) {
		Server := httptest.NewServer(gode.NewServer(gode.NewHub()))
		wsClient := mustDialWS(t, makeWebSocketURL(Server, "/casino/5145"))
		defer Server.Close()
		defer wsClient.Close()

		within(t, timeout, func() {
			assertReceiveBinaryMsg(t, wsClient, "5145")
		})
	})

	t.Run("/casino/5188 send ws msg 5145 on connect", func(t *testing.T) {
		Server := httptest.NewServer(gode.NewServer(gode.NewHub()))
		wsClient := mustDialWS(t, makeWebSocketURL(Server, "/casino/5188"))
		defer Server.Close()
		defer wsClient.Close()

		within(t, timeout, func() {
			assertReceiveBinaryMsg(t, wsClient, "5188")
		})
	})
}

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
		t.Errorf("message from web socket not matched\nwant %q\n got %q", want, got)
	}
}

func writeBinaryMsg(t *testing.T, wsClient *websocket.Conn, msg string) {
	err := wsClient.WriteMessage(websocket.BinaryMessage, []byte(msg))
	if err != nil {
		t.Error("ws WriteMessage Error", err)
	}
}

func within(t *testing.T, d time.Duration, assert func()) {
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
