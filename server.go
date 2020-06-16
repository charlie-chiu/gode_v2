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

	// for dev purpose: echo anything
	router.Handle("/echo", http.HandlerFunc(s.echoHandler))

	// handle game process
	router.Handle("/casino/", http.HandlerFunc(gameHandler))

	s.Handler = router

	return
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	gameType := strings.TrimLeft(r.URL.Path, "/casino/")

	fmt.Fprint(w, gameType)
}

func (s *Server) echoHandler(w http.ResponseWriter, r *http.Request) {
	client := &Client{
		IP: r.Header.Get("X-FORWARDED-FOR"),
	}
	_ = s.h.register(client)

	ws := newWSServer(w, r)
	_, p, _ := ws.ReadMessage()
	ws.writeBinaryMsg(p)
}
