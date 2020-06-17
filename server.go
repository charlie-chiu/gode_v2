package gode

import (
	"bytes"
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
	ws.writeBinaryMsg([]byte(`{"action":"ready"}`))

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
	action := s.parseClientAction(msg)
	switch action {
	case client.Login:
		ws.writeBinaryMsg([]byte(`{"action":"onLogin","result":{"event":"login"}}`))
		ws.writeBinaryMsg([]byte(`{"action":"onTakeMachine","result":{"event":"TakeMachine"}}`))
	case client.OnLoadInfo:
		ws.writeBinaryMsg([]byte(`{"action":"onOnLoadInfo2","result":{"event":"LoadInfo"}}`))
	case client.GetMachineDetail:
		ws.writeBinaryMsg([]byte(`{"action":"onGetMachineDetail","result":{"event":"MachineDetail"}}`))
	case client.BeginGame:
		ws.writeBinaryMsg([]byte(`{"action":"onBeginGame","result":{"event":"BeginGame"}}`))
	case client.ExchangeCredit:
		ws.writeBinaryMsg([]byte(`{"action":"onCreditExchange","result":{"event":"CreditExchange"}}`))
	case client.ExchangeBalance:
		ws.writeBinaryMsg([]byte(`{"action":"onBalanceExchange","result":{"event":"BalanceExchange"}}`))
	}
}

func (s *Server) parseClientAction(msg []byte) (action string) {
	//should refactor this
	if bytes.Contains(msg, []byte(client.Login)) {
		return client.Login
	}
	if bytes.Contains(msg, []byte(client.OnLoadInfo)) {
		return client.OnLoadInfo
	}
	if bytes.Contains(msg, []byte(client.GetMachineDetail)) {
		return client.GetMachineDetail
	}
	if bytes.Contains(msg, []byte(client.BeginGame)) {
		return client.BeginGame
	}
	if bytes.Contains(msg, []byte(client.ExchangeCredit)) {
		return client.ExchangeCredit
	}
	if bytes.Contains(msg, []byte(client.ExchangeBalance)) {
		return client.ExchangeBalance
	}

	return
}
