package gode_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gode"
)

type StubCaller struct{}

func (StubCaller) Call(service string, functionName string, parameters ...interface{}) ([]byte, error) {
	panic("implement me")
}

func TestMain(m *testing.M) {
	log.SetFlags(log.Lshortfile)

	os.Exit(m.Run())
}

func TestHub_NumberOfClients(t *testing.T) {
	hub := gode.NewHub()
	caller := &StubCaller{}
	Server := httptest.NewServer(gode.NewServer(hub, caller))
	defer Server.Close()

	t.Run("NumberOfClients return number of clients from hub", func(t *testing.T) {
		assertNumberOfClient(t, 0, hub.NumberOfClients())
	})

	client1 := mustDialWS(t, makeWebSocketURL(Server, "/casino/5145"))
	client2 := mustDialWS(t, makeWebSocketURL(Server, "/casino/5145"))
	_ = mustDialWS(t, makeWebSocketURL(Server, "/casino/5145"))

	t.Run("register when client connect", func(t *testing.T) {
		assertNumberOfClient(t, 3, hub.NumberOfClients())
	})

	t.Run("unregister when client connect", func(t *testing.T) {
		client1.Close()
		client2.Close()
		waitForProcess()
		assertNumberOfClient(t, 1, hub.NumberOfClients())
	})
}

func assertNumberOfClient(t *testing.T, wanted, got int) {
	t.Helper()
	if got != wanted {
		t.Errorf("wanted number of clients %d, got, %d", wanted, got)
	}
}

func TestRouter(t *testing.T) {
	t.Run("/ returns 404", func(t *testing.T) {
		caller := &StubCaller{}
		server := gode.NewServer(gode.NewHub(), caller)

		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, request)

		assertResponseCode(t, recorder.Code, http.StatusNotFound)
	})

	t.Run("get /casino/5145 returns 400 bad request", func(t *testing.T) {
		caller := &StubCaller{}
		server := gode.NewServer(gode.NewHub(), caller)

		request, _ := http.NewRequest(http.MethodGet, "/casino/5145", nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusBadRequest)
	})
}

func TestGameHandler(t *testing.T) {
	const timeout = 500 * time.Millisecond

	t.Run("not response when send incorrect data", func(t *testing.T) {
		caller := &StubCaller{}
		server := httptest.NewServer(gode.NewServer(gode.NewHub(), caller))
		client := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer client.Close()

		within(t, timeout, func() {
			assertReceiveBinaryMsg(t, client, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, client, `ola ola ola`)
		assertNoResponseWithin(t, timeout, client)
	})

	t.Run("not response when send incorrect action", func(t *testing.T) {
		caller := &StubCaller{}
		server := httptest.NewServer(gode.NewServer(gode.NewHub(), caller))
		client := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer client.Close()

		within(t, timeout, func() {
			assertReceiveBinaryMsg(t, client, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, client, `{"action": "hello"}`)
		assertNoResponseWithin(t, timeout, client)
	})

	t.Run("ws:/casino/5145 handle casino game process", func(t *testing.T) {
		caller := &StubCaller{}
		server := httptest.NewServer(gode.NewServer(gode.NewHub(), caller))
		client := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer client.Close()

		within(t, timeout, func() {
			//ready
			assertReceiveBinaryMsg(t, client, `{"action":"ready","result":null}`)

			//ClientLogin
			writeBinaryMsg(t, client, `{"action":"loginBySid","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
			assertReceiveBinaryMsg(t, client, `{"action":"onLogin","result":{"event":"login"}}`)
			assertReceiveBinaryMsg(t, client, `{"action":"onTakeMachine","result":{"event":"TakeMachine"}}`)

			//ClientOnLoadInfo
			writeBinaryMsg(t, client, `{"action":"onLoadInfo2","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
			assertReceiveBinaryMsg(t, client, `{"action":"onOnLoadInfo2","result":{"event":"LoadInfo"}}`)

			//ClientGetMachineDetail
			writeBinaryMsg(t, client, `{"action":"getMachineDetail","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
			assertReceiveBinaryMsg(t, client, `{"action":"onGetMachineDetail","result":{"event":"MachineDetail"}}`)

			//開分
			writeBinaryMsg(t, client, `{"action":"creditExchange"}`)
			assertReceiveBinaryMsg(t, client, `{"action":"onCreditExchange","result":{"event":"CreditExchange"}}`)

			//begin game
			writeBinaryMsg(t, client, `{"action":"beginGame4"}`)
			assertReceiveBinaryMsg(t, client, `{"action":"onBeginGame","result":{"event":"BeginGame"}}`)

			//洗分
			writeBinaryMsg(t, client, `{"action":"balanceExchange"}`)
			assertReceiveBinaryMsg(t, client, `{"action":"onBalanceExchange","result":{"event":"BalanceExchange"}}`)
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
		t.Errorf("message from web socket not matched\nwant %s\n got %s", want, got)
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
