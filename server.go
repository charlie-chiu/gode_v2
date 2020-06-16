package gode

import (
	"fmt"
	"net/http"
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

	router.Handle("/casino/5145", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, 5145)
	}))

	router.Handle("/casino/5156", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, 5156)
	}))

	s.Handler = router

	return
}
