package client

type WSData struct {
	Action    string `json:"action"`
	SessionID string `json:"sid"`
	BetBase   string `json:"rate"`
	Credit    string `json:"credit"`
	BetInfo   string `json:"betInfo"`
}
