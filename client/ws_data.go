package client

import "encoding/json"

type WSData struct {
	Action    string `json:"action"`
	SessionID string `json:"sid"`
	BetBase   string `json:"rate"`
	Credit    string `json:"credit"`
	BetInfo   string `json:"betInfo"`
}

type WSResponse struct {
	Action string          `json:"action"`
	Result json.RawMessage `json:"result"`
}
