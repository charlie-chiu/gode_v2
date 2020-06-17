package gode

import (
	"net/http"
	"strings"

	"gode/client"
)

type Server struct {
	http.Handler

	h *Hub
}

func NewServer(hub *Hub) (s *Server) {
	s = &Server{
		h: hub,
	}

	router := http.NewServeMux()
	// handle game process
	router.Handle("/casino/", http.HandlerFunc(s.gameHandler))
	s.Handler = router

	return
}

func (s *Server) gameHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := newWSServer(w, r)
	if err != nil {
		return
	}

	c := &client.Client{
		IP: r.Header.Get("X-FORWARDED-FOR"),
	}
	_ = s.h.register(c)

	//gameType := s.parseGameType(r)
	ws.writeBinaryMsg(client.Response(client.ReadyResponse, []byte(`null`)))

	// keep listen and handle ws messages
	wsMsg := make(chan []byte)
	go ws.listenJSON(wsMsg)
	for {
		closed := false
		select {
		case msg, ok := <-wsMsg:
			if ok {
				s.handleMessage(ws, msg)
			} else {
				//s.handleDisconnect()
				s.h.unregister(c)
				closed = true
			}
		}
		if closed {
			break
		}
	}
}

func (s *Server) parseGameType(r *http.Request) (gameType string) {
	return strings.TrimLeft(r.URL.Path, "/casino/")
}

func (s *Server) handleMessage(ws *wsServer, msg []byte) {
	action := client.ParseData(msg)
	switch action.Action {
	case client.Login:
		ws.writeBinaryMsg(client.Response(client.LoginResponse, []byte(`{"event":"login"}`)))
		ws.writeBinaryMsg(client.Response(client.TakeMachineResponse, []byte(`{"event":"TakeMachine"}`)))
	case client.OnLoadInfo:
		ws.writeBinaryMsg(client.Response(client.OnLoadInfoResponse, []byte(`{"event":"LoadInfo"}`)))
	case client.GetMachineDetail:
		ws.writeBinaryMsg(client.Response(client.GetMachineDetailResponse, []byte(`{"event":"MachineDetail"}`)))
	case client.BeginGame:
		ws.writeBinaryMsg(client.Response(client.BeginGameResponse, []byte(`{"event":"BeginGame"}`)))
	case client.ExchangeCredit:
		ws.writeBinaryMsg(client.Response(client.ExchangeCreditResponse, []byte(`{"event":"CreditExchange"}`)))
	case client.ExchangeBalance:
		ws.writeBinaryMsg(client.Response(client.ExchangeBalanceResponse, []byte(`{"event":"BalanceExchange"}`)))
	}
}
