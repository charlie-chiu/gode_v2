package gode_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"gode"
	"gode/client"
)

func TestMain(m *testing.M) {
	log.SetFlags(log.Lshortfile)

	os.Exit(m.Run())
}

func TestClient(t *testing.T) {
	t.Run("register on connect and unregister on disconnect", func(t *testing.T) {
		pool := gode.NewHub()
		Server := httptest.NewServer(gode.NewServer(pool, &SpyCaller{}))
		defer Server.Close()

		// no player
		assertNumberOfClient(t, 0, pool.NumberOfClients())

		player1 := mustDialWS(t, makeWebSocketURL(Server, "/casino/9999"))
		player2 := mustDialWS(t, makeWebSocketURL(Server, "/casino/9999"))
		player3 := mustDialWS(t, makeWebSocketURL(Server, "/casino/9999"))
		defer player3.Close()
		// 3 players
		assertNumberOfClient(t, 3, pool.NumberOfClients())

		player1.Close()
		player2.Close()
		waitForProcess()
		// 1 player
		assertNumberOfClient(t, 1, pool.NumberOfClients())
	})

	t.Run("store userID and hallID after loginBySID called", func(t *testing.T) {
		spyCaller := &SpyCaller{
			response: map[string]apiResponse{
				"loginCheck": apiResponse{
					result: []byte(`{"event":true, "data":{"user": {"UserID": "1325", "HallID":"0"}, "Session":{"Session":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}}}`),
					err:    nil,
				},
			},
		}
		spyHub := &SpyHub{}
		server := httptest.NewServer(gode.NewServer(spyHub, spyCaller))
		wsClient := mustDialWS(t, makeWebSocketURL(server, "/casino/5888"))
		defer server.Close()
		defer wsClient.Close()
		writeBinaryMsg(t, wsClient, `{"action":"loginBySid","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)

		waitForProcess()

		want := client.Client{
			GameType:  5888,
			UserID:    1325,
			HallID:    0,
			SessionID: "21d9b36e42c8275a4359f6815b859df05ec2bb0a",
		}
		got := *spyHub.clients[0]

		// just for testing, todo: remove this
		got.WSConn = nil

		// assert client equal
		if !reflect.DeepEqual(want, got) {
			t.Errorf("client not equal, \nwant: %+v\n got: %+v\n", want, got)
		}
	})
}

func TestRouter(t *testing.T) {
	t.Run("/ returns 404", func(t *testing.T) {
		caller := &SpyCaller{}
		server := gode.NewServer(gode.NewHub(), caller)

		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, request)

		assertResponseCode(t, recorder.Code, http.StatusNotFound)
	})

	t.Run("get /casino/5145 returns 400 bad request", func(t *testing.T) {
		caller := &SpyCaller{}
		server := gode.NewServer(gode.NewHub(), caller)

		request, _ := http.NewRequest(http.MethodGet, "/casino/5145", nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusBadRequest)
	})

	t.Run("/casino/{gameType} store game type in client", func(t *testing.T) {
		caller := &SpyCaller{}
		spyHub := &SpyHub{}
		server := httptest.NewServer(gode.NewServer(spyHub, caller))
		wsClient := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		wsClient2 := mustDialWS(t, makeWebSocketURL(server, "/casino/5188"))
		defer server.Close()
		defer wsClient.Close()
		defer wsClient2.Close()

		waitForProcess()
		if spyHub.clients[0].GameType != 5145 {
			t.Errorf("expected client0 has game type %d , got %d", 5145, spyHub.clients[0].GameType)
		}
		if spyHub.clients[1].GameType != 5188 {
			t.Errorf("expected client0 has game type %d , got %d", 5188, spyHub.clients[1].GameType)
		}
	})
}

func TestGameHandler(t *testing.T) {
	const timeout = 10 * time.Millisecond

	t.Run("not response when send incorrect data", func(t *testing.T) {
		caller := &SpyCaller{}
		server := httptest.NewServer(gode.NewServer(gode.NewHub(), caller))
		player := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer player.Close()

		assertWithin(t, timeout, func() {
			assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, player, `ola ola ola`)
		assertNoResponseWithin(t, timeout, player)
	})

	t.Run("not response when send incorrect action", func(t *testing.T) {
		caller := &SpyCaller{}
		server := httptest.NewServer(gode.NewServer(gode.NewHub(), caller))
		player := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer player.Close()

		assertWithin(t, timeout, func() {
			assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, player, `{"action": "hello"}`)
		assertNoResponseWithin(t, timeout, player)
	})
}

func TestProcess(t *testing.T) {
	const timeout = 10 * time.Millisecond

	callerResponses := map[string]apiResponse{
		"loginCheck": {
			result: []byte(`{"event":true, "data":{"user": {"UserID": "100", "HallID":"6"}, "Session":{"Session":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}}}`),
			err:    nil,
		},
	}

	// call api with correct parameter
	uid := uint32(100)
	hid := uint32(6)
	gameCode := uint16(0)
	sid := "21d9b36e42c8275a4359f6815b859df05ec2bb0a"
	betBase := "1:1"
	exchangeCredit := "50000"
	betInfo := json.RawMessage(`{"BetLevel":5}`)
	expectedHistory := apiHistory{
		{
			service:    "Client",
			function:   "loginCheck",
			parameters: []interface{}{sid},
		},
		{
			service:    "casino.slot.line243.BuBuGaoSheng",
			function:   "machineOccupy",
			parameters: []interface{}{uid, hid, gameCode},
		},
		{
			service:    "casino.slot.line243.BuBuGaoSheng",
			function:   "onLoadInfo",
			parameters: []interface{}{uid, gameCode},
		},
		{
			service:    "casino.slot.line243.BuBuGaoSheng",
			function:   "getMachineDetail",
			parameters: []interface{}{uid, gameCode},
		},
		{
			service:    "casino.slot.line243.BuBuGaoSheng",
			function:   "creditExchange",
			parameters: []interface{}{sid, gameCode, betBase, exchangeCredit},
		},
		{
			service:    "casino.slot.line243.BuBuGaoSheng",
			function:   "beginGame",
			parameters: []interface{}{sid, betInfo},
		},
		{
			service:    "casino.slot.line243.BuBuGaoSheng",
			function:   "balanceExchange",
			parameters: []interface{}{uid, hid, gameCode},
		},
	}

	spyCaller := &SpyCaller{response: callerResponses}
	server := httptest.NewServer(gode.NewServer(gode.NewHub(), spyCaller))
	player := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
	defer server.Close()
	defer player.Close()

	// response to player
	assertWithin(t, timeout, func() {
		//ready
		assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)

		//ClientLogin
		writeBinaryMsg(t, player, `{"action":"loginBySid","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
		assertReceiveBinaryMsg(t, player, `{"action":"onLogin","result":{"event":"login"}}`)
		assertReceiveBinaryMsg(t, player, `{"action":"onTakeMachine","result":{"event":"TakeMachine"}}`)

		//ClientOnLoadInfo
		writeBinaryMsg(t, player, `{"action":"onLoadInfo2","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
		assertReceiveBinaryMsg(t, player, `{"action":"onOnLoadInfo2","result":{"event":"LoadInfo"}}`)

		//ClientGetMachineDetail
		writeBinaryMsg(t, player, `{"action":"getMachineDetail","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
		assertReceiveBinaryMsg(t, player, `{"action":"onGetMachineDetail","result":{"event":"MachineDetail"}}`)

		//開分
		writeBinaryMsg(t, player, `{"action":"creditExchange","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a","rate":"1:1","credit":"50000"}`)
		assertReceiveBinaryMsg(t, player, `{"action":"onCreditExchange","result":{"event":"CreditExchange"}}`)

		//begin game
		writeBinaryMsg(t, player, `{"action":"beginGame4","sid":"123","betInfo":{"BetLevel":5}}`)
		assertReceiveBinaryMsg(t, player, `{"action":"onBeginGame","result":{"event":"BeginGame"}}`)

		//洗分
		writeBinaryMsg(t, player, `{"action":"balanceExchange"}`)
		assertReceiveBinaryMsg(t, player, `{"action":"onBalanceExchange","result":{"event":"BalanceExchange"}}`)
	})

	waitForProcess()

	assertLogEqual(t, expectedHistory, spyCaller.history)
}

func TestCasinoAPIErrorHandling(t *testing.T) {
	const timeout = 10 * time.Millisecond

	t.Run("disconnect when loginCheck error", func(t *testing.T) {
		caller := &SpyCaller{
			response: map[string]apiResponse{
				"loginCheck": {
					result: []byte(``),
					err:    fmt.Errorf("some api error"),
				},
			},
		}
		server := httptest.NewServer(gode.NewServer(gode.NewHub(), caller))
		player := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer player.Close()

		assertWithin(t, timeout, func() {
			assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, player, `{"action":"loginBySid","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
		assertNoResponseWithin(t, timeout, player)
	})

	t.Run("disconnect when machineOccupy error", func(t *testing.T) {
		caller := &SpyCaller{
			response: map[string]apiResponse{
				"machineOccupy": {
					result: []byte(``),
					err:    fmt.Errorf("some api error"),
				},
			},
		}
		server := httptest.NewServer(gode.NewServer(gode.NewHub(), caller))
		player := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer player.Close()

		assertWithin(t, timeout, func() {
			assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, player, `{"action":"loginBySid","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
		assertNoResponseWithin(t, timeout, player)
	})

	t.Run("disconnect when beginGame error", func(t *testing.T) {
		caller := &SpyCaller{
			response: map[string]apiResponse{
				"beginGame": {
					result: []byte(``),
					err:    fmt.Errorf("some api error"),
				},
			},
		}
		server := httptest.NewServer(gode.NewServer(gode.NewHub(), caller))
		player := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer player.Close()

		assertWithin(t, timeout, func() {
			assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, player, `{"action":"beginGame4","sid":"123","betInfo":{"BetLevel":5}}`)
		assertNoResponseWithin(t, timeout, player)
	})
}
