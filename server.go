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
				//handle msg
				if bytes.Contains(msg, []byte(`"action":"loginBySid"`)) {
					ws.writeBinaryMsg([]byte(`{"action":"onLogin","result":{"event":"login"}}`))
					ws.writeBinaryMsg([]byte(`{"action":"onTakeMachine","result":{"event":"TakeMachine"}}`))
				}
				if bytes.Contains(msg, []byte(`"action":"onLoadInfo2"`)) {
					ws.writeBinaryMsg([]byte(`{"action":"onOnLoadInfo2","result":{"event":"LoadInfo"}}`))
				}
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
