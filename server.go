package gode

import (
	"fmt"
	"net/http"
	"strings"
)

type Server struct {
	http.Handler
}

func NewServer() (s *Server) {
	s = &Server{}

	router := http.NewServeMux()

	// for dev purpose: echo anything
	router.Handle("/echo", http.HandlerFunc(echoHandler))

	// handle game process
	router.Handle("/casino/", http.HandlerFunc(gameHandler))

	s.Handler = router

	return
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	gameType := strings.TrimLeft(r.URL.Path, "/casino/")

	fmt.Fprint(w, gameType)
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
	ws := newWSServer(w, r)
	_, p, _ := ws.ReadMessage()
	ws.writeBinaryMsg(p)
}
