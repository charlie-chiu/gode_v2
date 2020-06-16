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
	ws, err := newWSServer(w, r)
	if err != nil {
		fmt.Fprint(w, "")
		return
	}

	gameType := strings.TrimLeft(r.URL.Path, "/casino/")
	ws.writeBinaryMsg([]byte(gameType))
}

func (s *Server) echoHandler(w http.ResponseWriter, r *http.Request) {
	client := &Client{
		IP: r.Header.Get("X-FORWARDED-FOR"),
	}
	_ = s.h.register(client)

	ws, _ := newWSServer(w, r)
	_, p, _ := ws.ReadMessage()
	ws.writeBinaryMsg(p)
}
