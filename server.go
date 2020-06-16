package gode

import "net/http"

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

	s.Handler = router

	return
}
