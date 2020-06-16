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

func TestHTTPRequest(t *testing.T) {
	t.Run("/ returns 404", func(t *testing.T) {
		server := gode.NewServer()

		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, request)

		assertResponseCode(t, recorder.Code, http.StatusNotFound)
	})

	t.Run("/casino/5145 return 5145", func(t *testing.T) {
		server := gode.NewServer()

		request, _ := http.NewRequest(http.MethodGet, "/casino/5145", nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)

		assertResponseCode(t, recorder.Code, http.StatusOK)
		s := recorder.Body.String()
		if s != "5145" {
			t.Errorf("want response body %s, got %s", "5145", s)
		}
	})

	t.Run("/casino/5156 return 5156", func(t *testing.T) {
		server := gode.NewServer()

		request, _ := http.NewRequest(http.MethodGet, "/casino/5156", nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)

		assertResponseCode(t, recorder.Code, http.StatusOK)
		s := recorder.Body.String()
		if s != "5156" {
			t.Errorf("want response body %s, got %s", "5156", s)
		}
	})
}

func TestWebSocket(t *testing.T) {
	const timeout = time.Second

	t.Run("/echo echo ws message before timeout", func(t *testing.T) {
		Server := httptest.NewServer(gode.NewServer())
		wsClient := mustDialWS(t, makeWebSocketURL(Server, "/echo"))
		defer Server.Close()
		defer wsClient.Close()

		within(t, timeout, func() {
			msg := "hello"
			writeBinaryMsg(t, wsClient, msg)
			assertReceiveBinaryMsg(t, wsClient, msg)
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
