package gode_test

import (
	"bytes"
	"fmt"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"gode/client"
	"gode/types"
)

type SpyCaller struct {
	history  apiHistory
	response map[string]apiResponse
	mutex    sync.Mutex
}

func (c *SpyCaller) Call(service types.GameType, function string, parameters ...interface{}) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.history = append(c.history, apiLog{
		service:    service,
		function:   function,
		parameters: parameters,
	})

	return c.response[function].result, c.response[function].err
}

type apiResponse struct {
	result []byte
	err    error
}

type apiHistory []apiLog

type apiLog struct {
	service    types.GameType
	function   string
	parameters []interface{}
}

func (l apiLog) String() string {
	b := strings.Builder{}

	b.WriteString(fmt.Sprintf("%v %v [", l.service, l.function))

	for _, parameter := range l.parameters {
		b.WriteString(fmt.Sprintf("(%T)%v ", parameter, parameter))
	}

	b.WriteString("]")

	return b.String()
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
	t.Helper()
	msgChan := make(chan []byte, 1)
	go func() {
		_, p, _ := client.ReadMessage()
		msgChan <- p
	}()

	select {
	case <-time.After(d):
		return
	case msg := <-msgChan:
		t.Errorf("shouldn't get response but got %s", msg)
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

func assertLogEqual(t *testing.T, expectedHistory, wantedLog apiHistory) {
	t.Helper()
	for i, expectedLog := range expectedHistory {
		if len(wantedLog) <= i {
			t.Fatalf("history %d not exists, want\n%+v\n", i, expectedLog)
		}
		gotLog := wantedLog[i]

		if !reflect.DeepEqual(gotLog, expectedLog) {
			t.Errorf("%dth api log not equal,\nwant:%v\n got:%v", i+1, expectedLog, gotLog)
		}
	}
}

func assertClientEqual(t *testing.T, want, got client.Client) {
	t.Helper()
	isGameTypeSame := want.GameType == got.GameType
	isUserIDSame := want.UserID == got.UserID
	isHallIDSame := want.HallID == got.HallID
	isSIDSame := bytes.Compare(want.SessionID, got.SessionID) == 0
	if !isGameTypeSame || !isUserIDSame || !isSIDSame || !isHallIDSame {
		t.Errorf("client not equal, \nwant: %+v\n got: %+v\n", want, got)
	}

}
