package client

import (
	"encoding/json"
	"log"
)

type Client struct {
	GameType uint16
}

func ParseData(msg []byte) *WSData {
	data := &WSData{}
	// already validate on listenJSON
	_ = json.Unmarshal(msg, data)

	return data
}

func Response(action string, result json.RawMessage) (data json.RawMessage) {
	response := &WSResponse{
		Action: action,
		Result: result,
	}

	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("client Response JSON Marshal error, %v", err)
	}

	return
}
