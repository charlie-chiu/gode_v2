package client

import "encoding/json"

type Client struct {
	IP string
}

func ParseData(msg []byte) *WSData {
	data := &WSData{}
	// already validate on listenJSON
	_ = json.Unmarshal(msg, data)

	return data
}
