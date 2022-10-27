package src

import (
	"727gpu_server/database"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"sort"
	"sync"
	"time"
)

type ReplyObj struct {
	Code int `json:"Code"`
	Type int `json:"Type"`
	Data any `json:"Data"`
}

type revObj struct {
	Type int                      `json:"Type"`
	Data []map[string]interface{} `json:"Data"`
}
type getObj struct { //type 2
	Target string `json:"Target"`
}

type infoObj struct { //type 1
	Ip   string             `json:"Ip"`
	Name string             `json:"Name"`
	Data []database.DataObj `json:"Data"`
}

var nodeConn = make(map[string]*websocket.Conn)
var portalConn = make(map[int]*websocket.Conn)

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

	var Ip string
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}
	defer func() {
		cancel()
		delete(nodeConn, Ip)
		closeSocketErr := ws.Close()
		if closeSocketErr != nil {
			panic(err)
		}
	}()
	err, handshakeMsg := HandShake(ws)
	Ip = handshakeMsg.Ip
	nodeConn[Ip] = ws
	if err != nil {
		_ = ws.WriteJSON(&ReplyObj{
			Code: 500,
			Data: "handshake failed!",
		})
		panic(err)
	}
	database.InsertMachine(db, handshakeMsg)

	wg.Add(1)
	go heartbeat(ctx, ws)

	for {
		var rev revObj

		err := ws.ReadJSON(&rev)
		if err != nil {
			panic(err)
		}
		if rev.Type == 1 {
			var msg []database.DataObj
			jsonStr, _ := json.Marshal(rev.Data)
			// Convert json string to struct
			if err := json.Unmarshal(jsonStr, &msg); err != nil {
				panic(err)
			}
			for _, d := range msg {
				database.InsertData(db, d)
			}
		}
		if rev.Type == 2 {
			var msg []database.ProcessObj
			jsonStr, _ := json.Marshal(rev.Data)
			// Convert json string to struct
			if err := json.Unmarshal(jsonStr, &msg); err != nil {
				panic(err)
			}
			for key := range portalConn {
				_ws := portalConn[key]
				_ws.WriteJSON(ReplyObj{Code: 200, Type: 2, Data: msg})
			}
		}
	}
	wg.Wait() //等待计数器置为0
}

func PortalHandler(c *gin.Context, db *sql.DB) {
	var wg sync.WaitGroup
	upGrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 5000 * time.Millisecond,
	}

	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	var Id int
	Id = len(portalConn)
	for {
		_, ok := portalConn[Id]
		if ok {
			Id++
		} else {
			break
		}
	}
	portalConn[Id] = ws
	if err != nil {
		panic(err)
	}
	ws.SetCloseHandler(func(code int, text string) error {
		wg.Done()
		fmt.Println("conn closed")
		return nil
	})
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go sendInfoLoop(ctx, db, ws)
	go sendMacInfoLoop(ctx, db, ws)

	wg.Wait() //等待计数器置为0
	defer func() {
		delete(portalConn, Id)
		cancel()
		closeSocketErr := ws.Close()
		if closeSocketErr != nil {
			panic(err)
		}
	}()

}

// type 0
func sendInfoLoop(ctx context.Context, db *sql.DB, ws *websocket.Conn) {
	lastTime := 0
LOOP:
	for {
		//fmt.Println(machines)
		//if ws.PingHandler() != nil {
		//	break
		//}
		machines, err := database.QueryAllMachine(db)
		if err != nil {
			fmt.Println(err)
		}
		for _, machine := range machines {
			dataList, err := database.QueryNewData(db, lastTime, machine.Ip)
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
				err = ws.WriteJSON(ReplyObj{Code: 200, Type: 0, Data: infoObj{Ip: machine.Ip, Name: machine.Name, Data: dataList}})
				if err != nil {
					fmt.Println(err)
				}

			}
		}
		select {
		case <-ctx.Done(): //等待通知
			break LOOP //跳出for循环
		default:

		}
		lastTime = int(time.Now().UnixNano() / 1e6)
		time.Sleep(2000 * time.Millisecond)
	}
}

// type 1
func sendMacInfoLoop(ctx context.Context, db *sql.DB, ws *websocket.Conn) {
LOOP:
	for {
		var rev revObj
		var msg []getObj
		err := ws.ReadJSON(&rev)
		if err != nil {
			fmt.Println(err)
			return
		}
		if rev.Type == 1 {
			jsonStr, _ := json.Marshal(rev.Data)
			// Convert json string to struct
			if err := json.Unmarshal(jsonStr, &msg); err != nil {
				fmt.Println(err)
				continue
			}
			target := msg[0].Target
			machine, err := database.QueryMachine(db, target)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = ws.WriteJSON(ReplyObj{Code: 200, Type: 1, Data: machine})
		}
		if rev.Type == 2 {
			jsonStr, _ := json.Marshal(rev.Data)
			// Convert json string to struct
			if err := json.Unmarshal(jsonStr, &msg); err != nil {
				fmt.Println(err)
				continue
			}
			if len(msg) > 0 {
				target := msg[0].Target
				_ws, ok := nodeConn[target]
				if ok {
					err = _ws.WriteJSON(ReplyObj{Code: 200, Type: 2, Data: 0})
				}
			}
		}
		select {
		case <-ctx.Done(): //等待通知
			break LOOP //跳出for循环
		default:

		}
	}
}

func heartbeat(ctx context.Context, ws *websocket.Conn) {
LOOP:
	for {
		ws.WriteJSON(ReplyObj{Code: 200, Type: -1, Data: 0})
		select {
		case <-ctx.Done(): //等待通知
			break LOOP //跳出for循环
		default:

		}
		time.Sleep(2000 * time.Millisecond)
	}
}
