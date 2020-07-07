package gode_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"gode"
	"gode/client"
	"gode/log"
	"gode/types"
)

// set up testing
func TestMain(m *testing.M) {
	log.SetLevel(log.Nothing)

	os.Exit(m.Run())
}

func TestClient(t *testing.T) {
	const LoginBySidMsg = `{"action":"loginBySid","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`
	const LoginAPIResult = `{"event":true, "data":{"user": {"UserID": "1325", "HallID":"10"}, "Session":{"Session":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}}}`

	t.Run("register client after login and unregister on disconnect", func(t *testing.T) {
		spyAPI := &SpyAPI{
			response: map[string]apiResponse{
				"loginCheck": {
					result: []byte(LoginAPIResult),
					err:    nil,
				},
			},
		}
		pool := gode.NewClientHub()
		Server := httptest.NewServer(gode.NewServer(pool, spyAPI))
		defer Server.Close()

		// no player
		assertNumberOfClient(t, 0, pool.NumberOfClients())

		player1 := mustDialWS(t, makeWebSocketURL(Server, "/casino/5100"))
		writeBinaryMsg(t, player1, LoginBySidMsg)
		player2 := mustDialWS(t, makeWebSocketURL(Server, "/casino/5200"))
		writeBinaryMsg(t, player2, LoginBySidMsg)
		player3 := mustDialWS(t, makeWebSocketURL(Server, "/casino/5300"))
		defer player3.Close()
		writeBinaryMsg(t, player3, LoginBySidMsg)

		// 3 players
		waitForProcess()
		assertNumberOfClient(t, 3, pool.NumberOfClients())

		// 1 player
		player1.Close()
		player2.Close()
		waitForProcess()
		assertNumberOfClient(t, 1, pool.NumberOfClients())
	})

	t.Run("store userID and hallID after loginBySID called", func(t *testing.T) {
		spyAPI := &SpyAPI{
			response: map[string]apiResponse{
				"loginCheck": {
					result: []byte(LoginAPIResult),
					err:    nil,
				},
			},
		}
		spyHub := &SpyHub{}
		server := httptest.NewServer(gode.NewServer(spyHub, spyAPI))
		player := mustDialWS(t, makeWebSocketURL(server, "/casino/5888"))
		defer server.Close()
		defer player.Close()
		writeBinaryMsg(t, player, LoginBySidMsg)

		waitForProcess()

		want := client.Client{
			GameType:  5888,
			UserID:    1325,
			HallID:    10,
			SessionID: types.SessionID("21d9b36e42c8275a4359f6815b859df05ec2bb0a"),
		}
		got := *spyHub.GetClient(0)

		assertClientEqual(t, want, got)
	})

	t.Run("/casino/{gameType} store game type in client", func(t *testing.T) {
		spyAPI := &SpyAPI{
			response: map[string]apiResponse{
				"loginCheck": {
					result: []byte(LoginAPIResult),
					err:    nil,
				},
			},
		}
		spyHub := &SpyHub{}
		server := httptest.NewServer(gode.NewServer(spyHub, spyAPI))
		player1 := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		writeBinaryMsg(t, player1, LoginBySidMsg)
		player2 := mustDialWS(t, makeWebSocketURL(server, "/casino/5188"))
		writeBinaryMsg(t, player2, LoginBySidMsg)
		defer server.Close()
		defer player1.Close()
		defer player2.Close()

		waitForProcess()
		if spyHub.GetClient(0).GameType != 5145 {
			t.Errorf("expected client0 has game type %d , got %d", 5145, spyHub.GetClient(0).GameType)
		}
		if spyHub.GetClient(1).GameType != 5188 {
			t.Errorf("expected client0 has game type %d , got %d", 5188, spyHub.GetClient(1).GameType)
		}
	})
}

func TestHandleClientException(t *testing.T) {
	const timeout = 10 * time.Millisecond

	t.Run("/ returns 404", func(t *testing.T) {
		spyAPI := &SpyAPI{}
		server := gode.NewServer(gode.NewClientHub(), spyAPI)

		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		recorder := httptest.NewRecorder()

		server.ServeHTTP(recorder, request)

		assertResponseCode(t, recorder.Code, http.StatusNotFound)
	})

	t.Run("get 404 not found when game type out of range", func(t *testing.T) {
		spyAPI := &SpyAPI{}
		server := gode.NewServer(gode.NewClientHub(), spyAPI)

		request, _ := http.NewRequest(http.MethodGet, "/casino/6666", nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusNotFound)
	})

	t.Run("get 404 not found when game type out of range", func(t *testing.T) {
		spyAPI := &SpyAPI{}
		server := gode.NewServer(gode.NewClientHub(), spyAPI)

		request, _ := http.NewRequest(http.MethodGet, "/casino/4999", nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusNotFound)
	})

	t.Run("get /casino/5145 returns 400 bad request", func(t *testing.T) {
		spyAPI := &SpyAPI{}
		server := gode.NewServer(gode.NewClientHub(), spyAPI)

		request, _ := http.NewRequest(http.MethodGet, "/casino/5145", nil)
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		assertResponseCode(t, recorder.Code, http.StatusBadRequest)
	})

	t.Run("not response when send incorrect ws data", func(t *testing.T) {
		spyAPI := &SpyAPI{}
		server := httptest.NewServer(gode.NewServer(gode.NewClientHub(), spyAPI))
		player := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer player.Close()

		assertWithin(t, timeout, func() {
			assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, player, `ola ola ola`)
		assertNoResponseWithin(t, timeout, player)
	})

	t.Run("not response when send incorrect ws action", func(t *testing.T) {
		spyAPI := &SpyAPI{}
		server := httptest.NewServer(gode.NewServer(gode.NewClientHub(), spyAPI))
		player := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer player.Close()

		assertWithin(t, timeout, func() {
			assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, player, `{"action": "hello"}`)
		assertNoResponseWithin(t, timeout, player)
	})

	t.Run("call leaveMachine when client disconnect", func(t *testing.T) {
		const timeout = 10 * time.Millisecond
		gameType := types.GameType(5199)
		svrPath := fmt.Sprintf("/casino/%d", gameType)

		spyAPI := &SpyAPI{response: map[string]apiResponse{
			"loginCheck": {
				result: []byte(`{"event":true, "data":{"user": {"UserID": "100", "HallID":"6"}, "Session":{"Session":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}}}`),
				err:    nil,
			},
			"balanceExchange": {
				result: []byte(`{"testing":"BalanceExchange"}`),
				err:    nil,
			},
			"machineLeave": {
				result: []byte(`{"testing":"LeaveMachine"}`),
				err:    nil,
			},
		}}
		server := httptest.NewServer(gode.NewServer(gode.NewClientHub(), spyAPI))
		player := mustDialWS(t, makeWebSocketURL(server, svrPath))
		defer server.Close()

		assertWithin(t, timeout, func() {
			assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, player, `{"action":"loginBySid","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)

		player.Close()

		sid := types.SessionID(`21d9b36e42c8275a4359f6815b859df05ec2bb0a`)
		uid := types.UserID(100)
		hid := types.HallID(6)
		gameCode := types.GameCode(0)
		expectedHistory := apiHistory{
			{
				service:    gameType,
				function:   "loginCheck",
				parameters: []interface{}{sid},
			},
			{
				service:    gameType,
				function:   "machineOccupy",
				parameters: []interface{}{uid, hid, gameCode},
			},
			{
				service:    gameType,
				function:   "balanceExchange",
				parameters: []interface{}{uid, hid, gameCode},
			},
			{
				service:    gameType,
				function:   "machineLeave",
				parameters: []interface{}{uid, hid, gameCode},
			},
		}
		waitForProcess()

		assertLogEqual(t, expectedHistory, spyAPI.History())
	})
}

func TestGameHandler(t *testing.T) {
	const timeout = 10 * time.Millisecond
	gameType := types.GameType(5199)
	svrPath := fmt.Sprintf("/casino/%d", gameType)

	spyAPI := &SpyAPI{response: map[string]apiResponse{
		"loginCheck": {
			result: []byte(`{"event":true, "data":{"user": {"UserID": "100", "HallID":"6"}, "Session":{"Session":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}}}`),
			err:    nil,
		},
		"machineOccupy": {
			result: []byte(`{"testing":"machineOccupy"}`),
			err:    nil,
		},
		"onLoadInfo": {
			result: []byte(`{"testing":"onLoadInfo"}`),
			err:    nil,
		},
		"getMachineDetail": {
			result: []byte(`{"testing":"getMachineDetail"}`),
			err:    nil,
		},
		"creditExchange": {
			result: []byte(`{"testing":"CreditExchange"}`),
			err:    nil,
		},
		"beginGame": {
			result: []byte(`{"testing":"BeginGame"}`),
			err:    nil,
		},
		"balanceExchange": {
			result: []byte(`{"testing":"BalanceExchange"}`),
			err:    nil,
		},
	}}
	server := httptest.NewServer(gode.NewServer(gode.NewClientHub(), spyAPI))
	player := mustDialWS(t, makeWebSocketURL(server, svrPath))
	defer server.Close()
	defer player.Close()

	t.Run("return casino api result within time", func(t *testing.T) {
		assertWithin(t, timeout, func() {
			//ready
			assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)

			//ClientLogin
			writeBinaryMsg(t, player, `{"action":"loginBySid","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
			assertReceiveBinaryMsg(t, player, `{"action":"onLogin","result":{"event":true,"data":{"user":{"UserID":"100","HallID":"6"},"Session":{"Session":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}}}}`)
			assertReceiveBinaryMsg(t, player, `{"action":"onTakeMachine","result":{"testing":"machineOccupy"}}`)

			//ClientOnLoadInfo
			writeBinaryMsg(t, player, `{"action":"onLoadInfo2","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
			assertReceiveBinaryMsg(t, player, `{"action":"onOnLoadInfo2","result":{"testing":"onLoadInfo"}}`)

			//ClientGetMachineDetail
			writeBinaryMsg(t, player, `{"action":"getMachineDetail","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
			assertReceiveBinaryMsg(t, player, `{"action":"onGetMachineDetail","result":{"testing":"getMachineDetail"}}`)

			//開分
			writeBinaryMsg(t, player, `{"action":"creditExchange","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a","rate":"1:1","credit":"50000"}`)
			assertReceiveBinaryMsg(t, player, `{"action":"onCreditExchange","result":{"testing":"CreditExchange"}}`)

			//begin game
			writeBinaryMsg(t, player, `{"action":"beginGame4","sid":"123","betInfo":{"BetLevel":5}}`)
			assertReceiveBinaryMsg(t, player, `{"action":"onBeginGame","result":{"testing":"BeginGame"}}`)

			//洗分
			writeBinaryMsg(t, player, `{"action":"balanceExchange"}`)
			assertReceiveBinaryMsg(t, player, `{"action":"onBalanceExchange","result":{"testing":"BalanceExchange"}}`)
		})
	})

	t.Run("called casino api with correct parameters", func(t *testing.T) {
		uid := types.UserID(100)
		hid := types.HallID(6)
		gameCode := types.GameCode(0)
		sid := types.SessionID(`21d9b36e42c8275a4359f6815b859df05ec2bb0a`)
		betBase := "1:1"
		exchangeCredit := types.Credit(50000)
		betInfo := types.BetInfo(`{"BetLevel":5}`)
		expectedHistory := apiHistory{
			{
				service:    gameType,
				function:   "loginCheck",
				parameters: []interface{}{sid},
			},
			{
				service:    gameType,
				function:   "machineOccupy",
				parameters: []interface{}{uid, hid, gameCode},
			},
			{
				service:    gameType,
				function:   "onLoadInfo",
				parameters: []interface{}{uid, gameCode},
			},
			{
				service:    gameType,
				function:   "getMachineDetail",
				parameters: []interface{}{uid, gameCode},
			},
			{
				service:    gameType,
				function:   "creditExchange",
				parameters: []interface{}{sid, gameCode, betBase, exchangeCredit},
			},
			{
				service:    gameType,
				function:   "beginGame",
				parameters: []interface{}{sid, gameCode, betInfo},
			},
			{
				service:    gameType,
				function:   "balanceExchange",
				parameters: []interface{}{uid, hid, gameCode},
			},
		}

		waitForProcess()
		assertLogEqual(t, expectedHistory, spyAPI.history)
	})
}

func TestHandleCasinoAPIException(t *testing.T) {
	const timeout = 10 * time.Millisecond

	t.Run("disconnect when loginCheck return invalid result", func(t *testing.T) {
		spyAPI := &SpyAPI{
			response: map[string]apiResponse{
				"loginCheck": {
					result: []byte(`oops`),
					err:    nil,
				},
			},
		}
		server := httptest.NewServer(gode.NewServer(gode.NewClientHub(), spyAPI))
		player := mustDialWS(t, makeWebSocketURL(server, "/casino/5145"))
		defer server.Close()
		defer player.Close()

		assertWithin(t, timeout, func() {
			assertReceiveBinaryMsg(t, player, `{"action":"ready","result":null}`)
		})

		writeBinaryMsg(t, player, `{"action":"loginBySid","sid":"21d9b36e42c8275a4359f6815b859df05ec2bb0a"}`)
		assertNoResponseWithin(t, timeout, player)
	})

	t.Run("disconnect when loginCheck error", func(t *testing.T) {
		spyAPI := &SpyAPI{
			response: map[string]apiResponse{
				"loginCheck": {
					result: []byte(``),
					err:    fmt.Errorf("some api error"),
				},
			},
		}
		server := httptest.NewServer(gode.NewServer(gode.NewClientHub(), spyAPI))
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
		spyAPI := &SpyAPI{
			response: map[string]apiResponse{
				"machineOccupy": {
					result: []byte(``),
					err:    fmt.Errorf("some api error"),
				},
			},
		}
		server := httptest.NewServer(gode.NewServer(gode.NewClientHub(), spyAPI))
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
		spyAPI := &SpyAPI{
			response: map[string]apiResponse{
				"beginGame": {
					result: []byte(``),
					err:    fmt.Errorf("some api error"),
				},
			},
		}
		server := httptest.NewServer(gode.NewServer(gode.NewClientHub(), spyAPI))
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
