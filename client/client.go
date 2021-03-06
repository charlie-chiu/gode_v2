package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"gode/log"
	"gode/types"
)

const messageType = websocket.BinaryMessage

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"gbcasino.bin"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	GameType  types.GameType
	UserID    types.UserID
	HallID    types.HallID
	SessionID types.SessionID

	WSConn *websocket.Conn
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
		log.Print(log.Notice, fmt.Sprintf("client Response JSON Marshal error, %v", err))
	}

	return
}

func (c *Client) ServeWS(w http.ResponseWriter, r *http.Request) error {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	c.WSConn = conn

	return nil
}

func (c *Client) ListenJSON(wsMsg chan []byte) {
	for {
		_, msg, err := c.WSConn.ReadMessage()
		if err != nil {
			log.Print(log.Notice, fmt.Sprintf("listenJSON ReadMessage Error: %v", err))
			close(wsMsg)
			break
		}

		//maybe shouldn't valid JSON here
		if !json.Valid(msg) {
			log.Print(log.Notice, fmt.Sprintf("listenJSON Valid JSON error, got %q", string(msg)))
			continue
		}

		wsMsg <- msg
	}
}

func (c *Client) WriteMsg(msg []byte) {
	err := c.WSConn.WriteMessage(messageType, msg)
	if err != nil {
		log.Print(log.Notice, fmt.Sprintf("WriteMsg Error: %v", err))
	}
}
