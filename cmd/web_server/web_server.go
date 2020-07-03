package main

import (
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"gode"
	"gode/casinoapi"
	"gode/log"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading .env file", err)
	}

	//set log level
	log.SetLevel(log.ParseLogLevel(os.Getenv("LOG_LEVEL")))

	hub := gode.NewClientHub()
	caller := casinoapi.NewFlash2db(os.Getenv("FLASH2DB_URL"))
	server := gode.NewServer(hub, caller)

	log.Fatal(http.ListenAndServe(":80", server))
}
