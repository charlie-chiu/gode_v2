package gode

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsServer struct {
	*websocket.Conn
}

func newWSServer(w http.ResponseWriter, r *http.Request) *wsServer {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Panicf("problem upgrading connection to web socket %v\n", err)
	}

	return &wsServer{conn}
}

func (w *wsServer) listenJSON(wsMsg chan []byte) {
	for {
		_, msg, err := w.ReadMessage()
		if err != nil {
			log.Println("listenJSON: ReadMessage Error: ", err)
			close(wsMsg)
			break
		}

		//maybe shouldn't valid JSON here
		if !json.Valid(msg) {
			log.Println("listenJSON: not Valid JSON", string(msg))
			continue
		}

		wsMsg <- msg
	}
}

func (w *wsServer) writeBinaryMsg(msg []byte) {
	err := w.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		log.Println("Write Error: ", err)
	}
}
