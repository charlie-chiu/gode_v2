package client

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const messageType = websocket.BinaryMessage

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	GameType  uint16
	UserID    uint32
	HallID    uint32
	SessionID string

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
		log.Printf("client Response JSON Marshal error, %v", err)
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
			//log.Printf("listenJSON ReadMessage Error: %v", err)
			close(wsMsg)
			break
		}

		//maybe shouldn't valid JSON here
		if !json.Valid(msg) {
			//log.Printf("listenJSON Valid JSON error, got %q", string(msg))
			continue
		}

		wsMsg <- msg
	}
}

func (c *Client) WriteMsg(msg []byte) {
	err := c.WSConn.WriteMessage(messageType, msg)
	if err != nil {
		log.Println("Write Error: ", err)
	}
}
