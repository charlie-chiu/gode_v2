package gode

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
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
		fmt.Fprint(w, "")
		return
	}

	client := &Client{
		IP: r.Header.Get("X-FORWARDED-FOR"),
	}
	_ = s.h.register(client)

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
				s.h.unregister(client)
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

const (
	// actions from client
	ClientLogin            = "loginBySid"
	ClientOnLoadInfo       = "onLoadInfo2"
	ClientGetMachineDetail = "getMachineDetail"
	ClientBeginGame        = "beginGame4"
	ClientExchangeCredit   = "creditExchange"
	ClientExchangeBalance  = "balanceExchange"
)

func (s *Server) handleMessage(ws *wsServer, msg []byte) {
	action := s.parseClientAction(msg)
	switch action {
	case ClientLogin:
		ws.writeBinaryMsg([]byte(`{"action":"onLogin","result":{"event":"login"}}`))
		ws.writeBinaryMsg([]byte(`{"action":"onTakeMachine","result":{"event":"TakeMachine"}}`))
	case ClientOnLoadInfo:
		ws.writeBinaryMsg([]byte(`{"action":"onOnLoadInfo2","result":{"event":"LoadInfo"}}`))
	case ClientGetMachineDetail:
		ws.writeBinaryMsg([]byte(`{"action":"onGetMachineDetail","result":{"event":"MachineDetail"}}`))
	case ClientBeginGame:
		ws.writeBinaryMsg([]byte(`{"action":"onBeginGame","result":{"event":"BeginGame"}}`))
	case ClientExchangeCredit:
		ws.writeBinaryMsg([]byte(`{"action":"onCreditExchange","result":{"event":"CreditExchange"}}`))
	case ClientExchangeBalance:
		ws.writeBinaryMsg([]byte(`{"action":"onBalanceExchange","result":{"event":"BalanceExchange"}}`))
	}
}

func (s *Server) parseClientAction(msg []byte) (action string) {
	//should refactor this
	if bytes.Contains(msg, []byte(ClientLogin)) {
		return ClientLogin
	}
	if bytes.Contains(msg, []byte(ClientOnLoadInfo)) {
		return ClientOnLoadInfo
	}
	if bytes.Contains(msg, []byte(ClientGetMachineDetail)) {
		return ClientGetMachineDetail
	}
	if bytes.Contains(msg, []byte(ClientBeginGame)) {
		return ClientBeginGame
	}
	if bytes.Contains(msg, []byte(ClientExchangeCredit)) {
		return ClientExchangeCredit
	}
	if bytes.Contains(msg, []byte(ClientExchangeBalance)) {
		return ClientExchangeBalance
	}

	return
}
