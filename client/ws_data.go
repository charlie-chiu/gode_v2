package client

import (
	"encoding/json"

	"gode/types"
)

type WSData struct {
	Action    string          `json:"action"`
	SessionID types.SessionID `json:"sid"`
	BetBase   string          `json:"rate"`
	Credit    string          `json:"credit"`
	BetInfo   json.RawMessage `json:"betInfo"`
}

type WSResponse struct {
	Action string          `json:"action"`
	Result json.RawMessage `json:"result"`
}
