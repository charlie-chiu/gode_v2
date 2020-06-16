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
	router.Handle("/echo", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ws := newWSServer(w, r)
		_, p, _ := ws.ReadMessage()
		ws.writeBinaryMsg(p)
	}))

	router.Handle("/casino/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gameType := strings.TrimLeft(r.URL.Path, "/casino/")

		fmt.Fprint(w, gameType)
	}))

	s.Handler = router

	return
}
