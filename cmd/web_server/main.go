package main

import (
	"log"
	"net/http"

	"gode"
	"gode/casinoapi"
)

func main() {
	hub := gode.NewClientHub()
	caller := casinoapi.NewFlash2db("http://103.241.238.74/")
	server := gode.NewServer(hub, caller)

	log.Fatal(http.ListenAndServe(":80", server))
}
