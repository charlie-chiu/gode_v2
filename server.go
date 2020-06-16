package gode

import (
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

	gameType := s.parseGameType(r)
	ws.writeBinaryMsg([]byte(gameType))

	// handle disconnect
	_, _, err = ws.ReadMessage()
	if err != nil {
		s.h.unregister(client)
	}
}

func (s *Server) parseGameType(r *http.Request) (gameType string) {
	gameType = strings.TrimLeft(r.URL.Path, "/casino/")

	return
}
