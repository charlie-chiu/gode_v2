package main

import (
	"log"
	"net/http"

	"gode"
)

func main() {
	hub := gode.NewHub()
	caller := &StubAPI{}
	server := gode.NewServer(hub, caller)

	log.Fatal(http.ListenAndServe(":80", server))
}

type StubAPI struct{}

func (StubAPI) Call(service string, functionName string, parameters ...interface{}) ([]byte, error) {
	panic("implement me")
}
