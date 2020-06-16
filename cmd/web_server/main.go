package main

import (
	"log"
	"net/http"

	"gode"
)

func main() {
	hub := gode.NewHub()
	server := gode.NewServer(hub)

	log.Fatal(http.ListenAndServe(":80", server))
}
