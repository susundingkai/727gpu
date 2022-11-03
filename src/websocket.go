package src

import (
	"727gpu_server/config"
	"727gpu_server/database"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
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
	Id     int    `json:"Id"`
	Code   string `json:"Code"`
	MemTH  int    `json:"MemTH"`
}

type infoObj struct { //type 1
	Ip   string             `json:"Ip"`
	Name string             `json:"Name"`
	Data []database.DataObj `json:"Data"`
}
type wxRevObj struct {
	openid       string `json:"openid"`
	session_key  string `json:"session_key"`
	unionid      string `json:"unionid"`
	errcode      int    `json:"errcode"`
	errmsg       string `json:"errmsg"`
	access_token string `json:"access_token"`
	expires_in   int    `json:"expires_in"`
}

var myConfig = config.ReadConfig()
var nodeConn = make(map[string]*websocket.Conn)
var portalConn = make(map[int]*websocket.Conn)
var subGPU = make(map[string]map[getObj]int) // [ip]->{target ...}
var code2session = make(map[string]string)

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
func NodeHandler(c *gin.Context, db *sql.DB) {

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
		for obj, _ := range subGPU[Ip] {
			sendFailSubscribe(Ip, obj)
		}
		for obj, _ := range portalConn {
			_ws := portalConn[obj]
			_ws.WriteJSON(ReplyObj{Code: 200, Type: 1, Data: Ip}) //发送离线信息
		}
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
		if rev.Type == 1 { //收到gpu信息
			var msg []database.DataObj
			jsonStr, _ := json.Marshal(rev.Data)
			// Convert json string to struct
			if err := json.Unmarshal(jsonStr, &msg); err != nil {
				panic(err)
			}
			for _, d := range msg {
				ok, target, obj := compareSub(d, subGPU)
				if ok {
					sendSubscribe(target, obj)
					delete(subGPU[target], obj)
				}
				database.InsertData(db, d)
			}
		}
		if rev.Type == 2 { //收到proc信息
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
	go revGpuInfoLoop(ctx, db, ws)
	go portalRevLoop(ctx, db, ws)

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
func revGpuInfoLoop(ctx context.Context, db *sql.DB, ws *websocket.Conn) {
	lastTime := int(time.Now().UnixNano() / 1e6)
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
func portalRevLoop(ctx context.Context, db *sql.DB, ws *websocket.Conn) {
LOOP:
	for {
		var rev revObj
		var msg []getObj
		err := ws.ReadJSON(&rev)
		if err != nil {
			fmt.Println(err)
			return
		}
		if rev.Type == 1 { //查询机器
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
		if rev.Type == 2 { //请求PROC

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
		if rev.Type == 3 { //请求订阅

			jsonStr, _ := json.Marshal(rev.Data)
			// Convert json string to struct
			fmt.Println("订阅信息:", string(jsonStr))
			if err := json.Unmarshal(jsonStr, &msg); err != nil {
				fmt.Println(err)
				continue
			}
			if len(msg) > 0 {
				var session string
				var ok bool
				if session, ok = code2session[msg[0].Code]; ok {
					fmt.Println("get session:", session)
				} else {
					ok, session = getSession(msg)
					if ok == false {
						continue
					}
					code2session[msg[0].Code] = session
				}
				msg[0].Code = session
				if subGPU[msg[0].Target] == nil {
					subGPU[msg[0].Target] = make(map[getObj]int)
				}
				subGPU[msg[0].Target][msg[0]] = int(time.Now().UnixNano() / 1e6)
				fmt.Println("订阅成功")
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
		err := ws.WriteJSON(ReplyObj{Code: 200, Type: -1, Data: 0})
		if err != nil {
			return
		}
		select {
		case <-ctx.Done(): //等待通知
			break LOOP //跳出for循环
		default:

		}
		time.Sleep(2000 * time.Millisecond)
	}
}

func compareSub(newInfo database.DataObj, curSub map[string]map[getObj]int) (bool, string, getObj) {
	if _, ok := curSub[newInfo.Ip]; ok {
		for k, _ := range curSub[newInfo.Ip] {
			if k.Id == newInfo.GpuId && k.MemTH > int((newInfo.MemUsed/newInfo.MemTotal)*100) {
				return true, newInfo.Ip, k
			}
		}
	}
	return false, "", getObj{}
}
func getSession(msg []getObj) (bool, string) {
	var revObj map[string]interface{}
	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", myConfig.Wx.Appid, myConfig.Wx.Appsecret, msg[0].Code)
	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false, ""
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)

	}
	if err := json.Unmarshal(body, &revObj); err != nil {
		fmt.Println(err)
		return false, ""
	}
	if revObj["errmsg"] != nil {
		fmt.Println(revObj["errcode"], revObj["errmsg"], msg[0].Code)
		return false, ""
	}
	return true, revObj["openid"].(string)
}
func sendSubscribe(target string, msg getObj) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", myConfig.Wx.Appid, myConfig.Wx.Appsecret)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var revObj map[string]interface{}
	if err := json.Unmarshal(body, &revObj); err != nil {
		fmt.Println(err)
		return
	}
	if revObj["errmsg"] != nil {
		return
	}
	accessToken := revObj["access_token"]
	url = fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", accessToken)
	method = "POST"
	payload := strings.NewReader(fmt.Sprintf(`{
    "touser": "%s",
    "template_id": "%s",
    "data": {
        "thing13": {
            "value": "%s"
        },
        "time19": {
            "value": "%s"
        },
        "phrase12": {
            "value": "%s"
        }
    }
}`, msg.Code, myConfig.Wx.SubSucee, fmt.Sprintf("%s:%d", target, msg.Id), time.Now().Format("2006-01-02 15:04:05"), "空闲"))
	client = &http.Client{}
	req, err = http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
func sendFailSubscribe(target string, msg getObj) {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", myConfig.Wx.Appid, myConfig.Wx.Appsecret)
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var revObj map[string]interface{}
	if err := json.Unmarshal(body, &revObj); err != nil {
		fmt.Println(err)
		return
	}
	if revObj["errmsg"] != nil {
		return
	}
	accessToken := revObj["access_token"]
	url = fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", accessToken)
	method = "POST"
	payload := strings.NewReader(fmt.Sprintf(`{
    "touser": "%s",
    "template_id": "%s",
    "data": {
        "thing5": {
            "value": "%s"
        },
        "time2": {
            "value": "%s"
        },
        "thing3": {
            "value": "%s"
        }
    }

}`, msg.Code, myConfig.Wx.SubFailed, fmt.Sprintf("%s:%d", target, msg.Id), time.Unix(int64(subGPU[target][msg]/1000), 0).Format("2006-01-02 15:04:05"), "服务器离线"))
	client = &http.Client{}
	req, err = http.NewRequest(method, url, payload)
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err = client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
