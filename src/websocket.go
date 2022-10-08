package src

import (
	"727gpu_server/database"
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

type ReplyObj struct {
	Code int `json:"Code"`
	Data any `json:"Data"`
}

type revObj struct {
	Type int                    `json:"Type"`
	Data map[string]interface{} `json:"Data"`
}

func HandShake(ws *websocket.Conn) (error, database.MachineObj) {
	var rev revObj
	var firstMsg database.MachineObj
	err := ws.ReadJSON(&rev)
	if err != nil {
		panic(err)
	}
	if rev.Type == 0 {
		jsonStr, _ := json.Marshal(rev.Data)
		// Convert json string to struct
		if err := json.Unmarshal(jsonStr, &firstMsg); err != nil {
			panic(err)
		}
	}
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
		var rev revObj
		var msg database.DataObj
		err := ws.ReadJSON(&rev)
		if err != nil {
			panic(err)
		}
		if rev.Type == 1 {
			jsonStr, _ := json.Marshal(rev.Data)
			// Convert json string to struct
			if err := json.Unmarshal(jsonStr, &msg); err != nil {
				panic(err)
			}
			database.InsertData(db, msg)
		}
	}
}
