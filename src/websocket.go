package src

import (
	"727gpu_server/database"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

type ReplyObj struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

func HandShake(ws *websocket.Conn) (error, database.MachineObj) {
	var firstMsg database.MachineObj
	err := ws.ReadJSON(&firstMsg)
	return err, firstMsg
}

func SocketHandler(c *gin.Context, db *sql.DB) {
	upGrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}
	defer func() {
		closeSocketErr := ws.Close()
		if closeSocketErr != nil {
			panic(err)
		}
	}()
	err, handshakeMsg := HandShake(ws)
	if err != nil {
		_ = ws.WriteJSON(&ReplyObj{
			Code: 500,
			Data: "handshake failed!",
		})
		panic(err)
	}
	database.InsertMachine(db, handshakeMsg)
	for {
		msgType, msg, err := ws.ReadMessage()
		if err != nil {
			//panic(err)
			break
		}
		fmt.Printf("Message Type: %d, Message: %s\n", msgType, string(msg))
		err = ws.WriteJSON(struct {
			Reply string `json:"reply"`
		}{
			Reply: "Echo...",
		})
		if err != nil {
			//panic(err)
			break
		}
	}
}
