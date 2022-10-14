package src

import (
	"727gpu_server/database"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sort"
	"time"
)

type ReplyObj struct {
	Code int `json:"Code"`
	Data any `json:"Data"`
}

type revObj struct {
	Type int                      `json:"Type"`
	Data []map[string]interface{} `json:"Data"`
}
type sendInfoObj struct {
	Code int     `json:"Code"`
	Data infoObj `json:"Data"`
}

type infoObj struct {
	Ip   string             `json:"Ip"`
	Name string             `json:"Name"`
	Data []database.DataObj `json:"Data"`
}

func HandShake(ws *websocket.Conn) (error, database.MachineObj) {
	var rev revObj
	var firstMsg database.MachineObj
	err := ws.ReadJSON(&rev)
	if err != nil {
		panic(err)
	}
	if rev.Type == 0 {
		jsonStr, _ := json.Marshal(rev.Data[0])
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
		var msg []database.DataObj
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
			for _, d := range msg {
				database.InsertData(db, d)
			}
		}
	}
}

func ProtalHandler(c *gin.Context, db *sql.DB) {
	upGrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 5000 * time.Millisecond,
	}

	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	lastTime := 0
	if err != nil {
		panic(err)
	}
	defer func() {
		closeSocketErr := ws.Close()
		if closeSocketErr != nil {
			panic(err)
		}
	}()
	for {
		//fmt.Println(machines)
		//if ws.PingHandler() != nil {
		//	break
		//}
		machines, err := database.QueryAllMachine(db)

		if err != nil {
			panic(err)
		}
		for _, machine := range machines {
			fmt.Println(machine)
			dataList, err := database.QueryNewData(db, lastTime, machine.Ip)
			fmt.Println(dataList)
			if err != nil {
				fmt.Println(err)
				continue
			}
			sort.Sort(dataList)
			length := dataList.Len()
			if length == 0 {
				continue
			}
			if dataList[length-1].Time > lastTime {
				err = ws.WriteJSON(ReplyObj{Code: 200, Data: infoObj{Ip: machine.Ip, Name: machine.Name, Data: dataList}})
				if err != nil {
					fmt.Println(err)
					panic(err)
				}

			}
		}
		lastTime = int(time.Now().UnixNano() / 1e6)
		time.Sleep(5000 * time.Millisecond)
	}
}
