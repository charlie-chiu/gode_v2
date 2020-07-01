package gode

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"gode/casinoapi"
	"gode/client"
	"gode/types"
)

type Server struct {
	http.Handler

	clients ClientPool

	api casinoapi.Caller
}

func NewServer(clients ClientPool, caller casinoapi.Caller) (s *Server) {
	s = &Server{
		clients: clients,
		api:     caller,
	}

	router := http.NewServeMux()
	// handle game process
	router.Handle("/casino/", http.HandlerFunc(s.gameHandler))
	s.Handler = router

	return
}

func (s *Server) gameHandler(w http.ResponseWriter, r *http.Request) {
	c := &client.Client{}
	err := c.ServeWS(w, r)
	if err != nil {
		return
	}

	// make sure every connection will get different client
	gameType, _ := s.parseGameType(r)
	c.GameType = gameType
	_ = s.clients.Register(c)

	c.WriteMsg(client.Response(client.ReadyResponse, []byte(`null`)))

	// keep listen and handle ws messages
	wsMsg := make(chan []byte)
	go c.ListenJSON(wsMsg)
	for {
		closed := false
		select {
		case msg, ok := <-wsMsg:
			if ok {
				s.handleMessage(msg, c)
			} else {
				//s.handleDisconnect()
				s.clients.Unregister(c)
				closed = true
			}
		}
		if closed {
			break
		}
	}
}

func (s *Server) parseGameType(r *http.Request) (gameType types.GameType, err error) {
	gameTypeStr := strings.TrimLeft(r.URL.Path, "/casino/")
	gameTypeUint64, err := strconv.ParseUint(gameTypeStr, 10, 0)

	return types.GameType(gameTypeUint64), err
}

type LoginCheckResult struct {
	Event bool `json:"event"`
	Data  struct {
		Session struct {
			Session  types.SessionID `json:"session"`
			CreateAt string          `json:"create_at"`
		} `json:"session"`
		User struct {
			UserID       types.UserID `json:"UserID"`
			Username     string       `json:"Username"`
			LoginName    string       `json:"LoginName"`
			Currency     string       `json:"Currency"`
			Cash         string       `json:"Cash"`
			HallID       types.HallID `json:"HallID"`
			ExchangeRate string       `json:"ExchangeRate"`
			Test         string       `json:"Test"`
		} `json:"user"`
	} `json:"data"`
}

func (s *Server) handleMessage(msg []byte, c *client.Client) {
	dummyGameCode := types.GameCode(0)
	data := client.ParseData(msg)

	switch data.Action {
	case client.Login:
		loginCheckResult, err := s.api.Call("Client", "loginCheck", data.SessionID)
		if err != nil {
			return
		}

		if err := storeLoginResult(loginCheckResult, c); err != nil {
			return
		}

		apiResult, err := s.api.Call("casino.slot.line243.BuBuGaoSheng", "machineOccupy", c.UserID, c.HallID, dummyGameCode)
		if err != nil {
			return
		}

		c.WriteMsg(client.Response(client.LoginResponse, []byte(`{"event":"login"}`)))
		c.WriteMsg(client.Response(client.TakeMachineResponse, apiResult))

	case client.OnLoadInfo:
		apiResult, _ := s.api.Call("casino.slot.line243.BuBuGaoSheng", "onLoadInfo", c.UserID, dummyGameCode)
		c.WriteMsg(client.Response(client.OnLoadInfoResponse, apiResult))

	case client.GetMachineDetail:
		apiResult, _ := s.api.Call("casino.slot.line243.BuBuGaoSheng", "getMachineDetail", c.UserID, dummyGameCode)
		c.WriteMsg(client.Response(client.GetMachineDetailResponse, apiResult))

	case client.BeginGame:
		apiResult, err := s.api.Call("casino.slot.line243.BuBuGaoSheng", "beginGame", c.SessionID, data.BetInfo)
		if err != nil {
			return
		}
		c.WriteMsg(client.Response(client.BeginGameResponse, apiResult))

	case client.ExchangeCredit:
		apiResult, _ := s.api.Call("casino.slot.line243.BuBuGaoSheng", "creditExchange", c.SessionID, dummyGameCode, data.BetBase, data.Credit)
		c.WriteMsg(client.Response(client.ExchangeCreditResponse, apiResult))

	case client.ExchangeBalance:
		apiResult, _ := s.api.Call("casino.slot.line243.BuBuGaoSheng", "balanceExchange", c.UserID, c.HallID, dummyGameCode)
		c.WriteMsg(client.Response(client.ExchangeBalanceResponse, apiResult))
	}
}

func storeLoginResult(loginCheckResult []byte, c *client.Client) error {
	result := &LoginCheckResult{}
	err := json.Unmarshal(loginCheckResult, result)
	if err != nil {
		return err
	}
	c.HallID = result.Data.User.HallID
	c.UserID = result.Data.User.UserID
	c.SessionID = result.Data.Session.Session

	return nil
}
