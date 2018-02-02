package main

import (
	"log"
	"net/http"
	"time"
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/bwmarrin/snowflake"
	"strconv"
	"strings"
)

//localhost
const (
	dbuser     = "goconn"
	dbpassword = "123456"
	dbhost     = "localhost"
	dbport     = "3306"
	dbname     = "eth"
)

const (
	sevName = "TestServer"
)

var msgCode = make(map[int64]bool)

type ChatMsg struct {
	ID         int       `gorm:"primary_key"`
	Type       int       `gorm:"column:type"`
	FromSccode string    `gorm:"column:from_sccode"`
	ToSccode   string    `gorm:"column:to_sccode"`
	Contact    string    `gorm:"column:contact"`
	IsRead     int       `gorm:"column:is_read"`
	MsgCode    int64     `gorm:"column:msg_code"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (ChatMsg) TableName() string {
	return "chat_msg"
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 204800
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	sccode string
}

type Msg struct {
	From string
	To   string
	Msg  string
	Code string
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func servec2Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/c2" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "home-c2.html")
}

func servesHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/s" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "home-s.html")
}

func servesDoRead(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/doRead" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	msgCode := r.PostFormValue("msgCode")

	if msgCode == "" {
		var rs = "error"
		b := []byte(rs)
		w.Write(b)
	} else {
		arr := strings.Split(msgCode, ",")
		for _, code := range arr {
			if code == "" {

			} else {
				msgCodeInt, err := strconv.ParseInt(code, 10, 64)
				if err != nil {
					log.Printf("error: %v", err)
				} else {
					go doRead(msgCodeInt, 10)
				}
			}
		}

		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
		var rs = "succes"
		b := []byte(rs)
		w.Write(b)
	}

}

//func servesGetRecord(w http.ResponseWriter, r *http.Request) {
//	if r.URL.Path != "/getRecord" {
//		http.Error(w, "Not found", 404)
//		return
//	}
//	if r.Method != "POST" {
//		http.Error(w, "Method not allowed", 405)
//		return
//	}
//
//	scCode := r.PostFormValue("scCode")
//
//	var args = ""
//	args += dbuser
//	args += ":"
//	args += dbpassword
//	args += "@("
//	args += dbhost
//	if dbport != "" {
//		args += ":"
//		args += dbport
//	}
//	args += ")/"
//	args += dbname
//	args += "?charset=utf8&parseTime=True&loc=Local"
//
//	//db, err := gorm.Open("mysql", "root:@/eth?charset=utf8&parseTime=True&loc=Local")
//	db, err := gorm.Open("mysql", args)
//	if err != nil {
//		log.Printf("error: %v", err)
//	}
//	defer db.Close()
//
//	//chatMsg := ChatMsg{}
//
//	w.Header().Set("Cache-Control", "no-cache")
//	w.Header().Set("Content-Type", "application/json")
//
//	rows, err := db.Model(&ChatMsg{}).
//		Where("from_sccode = ?", scCode).
//		Or("to_sccode = ?", scCode).
//		Order("created_at desc, id desc").
//		Offset(0).
//		Limit(1000).
//		Rows()
//	defer rows.Close()
//	if err != nil {
//		log.Printf("error: %v", err)
//		var rs = "[]"
//		b := []byte(rs)
//		w.Write(b)
//	} else {
//		rs := make([]ChatMsg, 0)
//
//		for rows.Next() {
//			var user ChatMsg
//			db.ScanRows(rows, &user)
//			rs = append(rs, user)
//		}
//
//		b, err := json.Marshal(rs)
//		if err != nil {
//			log.Printf("error: %v", err)
//		}
//
//		w.Write(b)
//	}
//
//}

func main() {

	hub := newHub()
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/s", servesHome)
	http.HandleFunc("/c2", servec2Home)
	http.HandleFunc("/doRead", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		servesDoRead(w, r)
	})
	//http.HandleFunc("/getRecord", servesGetRecord)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err2 := http.ListenAndServe(":8080", nil)
	if err2 != nil {
		log.Fatal("ListenAndServe: ", err2)
	}

}

// hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				//select {
				//case client.send <- message:
				//default:
				//	close(client.send)
				//	delete(h.clients, client)
				//}

				str5 := string(message[:])
				stb := &Msg{}
				err := json.Unmarshal([]byte(str5), &stb)
				if err != nil {
					log.Printf("error: %v", err)
				}
				//接收者是否存在
				var hasit = 0
				if stb.To == "sev" {
					if client.sccode == sevName {
						hasit = 1
					}
				} else {
					if client.sccode == stb.To {
						hasit = 1
					}
				}

				//fmt.Print(hasit)

				if hasit == 1 {
					client.send <- message
				}

			}
		}
	}
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump(hub *Hub) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		//message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		//c.hub.broadcast <- message

		str3 := string(message[:])

		stb := &Msg{}
		err = json.Unmarshal([]byte(str3), &stb)
		if err != nil {
			log.Printf("error: %v", err)
		}

		//接收者是否存在
		hasit := 0
		for k := range hub.clients {
			if stb.To == "sev" {
				if k.sccode == sevName {
					hasit = 1
				}
			} else {
				if k.sccode == stb.To {
					hasit = 1
				}
			}
		}

		fromSccode := stb.From
		toSccode := stb.To

		if fromSccode == "sev" {
			fromSccode = sevName
		}

		if toSccode == "sev" {
			toSccode = sevName
		}

		var msgCodeStr = time.Now().UnixNano()
		msgCodeStr2 := strconv.FormatInt(msgCodeStr, 10)
		node, err := snowflake.NewNode(8)
		if err != nil {
			log.Printf("error: %v", err)
		} else {
			msgCodeStrId := node.Generate()
			msgCodeStr2 = msgCodeStrId.String()
		}
		stb.Code = msgCodeStr2

		newMessage, err := json.Marshal(stb)
		if err != nil {
			log.Printf("error: %v", err)
			var isread = 0
			if hasit == 1 {
				c.hub.broadcast <- newMessage
				isread = 1
			} else {
				isread = 0
			}
			go msgStore(1, fromSccode, toSccode, stb.Msg, isread, msgCodeStr2)
		} else {
			var isread = 0
			if hasit == 1 {
				c.hub.broadcast <- newMessage
				isread = 1
			} else {
				isread = 0
			}
			go msgStore(1, fromSccode, toSccode, stb.Msg, isread, msgCodeStr2)
		}

	}
}

func doRead(msgCodes int64, times int) {

	if times <= 0 {
		return
	}

	var args = ""
	args += dbuser
	args += ":"
	args += dbpassword
	args += "@("
	args += dbhost
	if dbport != "" {
		args += ":"
		args += dbport
	}
	args += ")/"
	args += dbname
	args += "?charset=utf8&parseTime=True&loc=Local"

	//db, err := gorm.Open("mysql", "root:@/eth?charset=utf8&parseTime=True&loc=Local")
	db, err := gorm.Open("mysql", args)
	if err != nil {
		log.Printf("error: %v", err)
	}
	defer db.Close()

	chatMsg := ChatMsg{}

	msgCodesStr := strconv.FormatInt(msgCodes, 10)
	if err := db.Where("msg_code = ?", msgCodesStr).First(&chatMsg).Error; err != nil {
		//log.Printf("error: %v", err)

		time.Sleep(2 * time.Second)
		times = times - 1
		go doRead(msgCodes, times)
	} else {
		if db.Where("msg_code = ?", msgCodesStr).First(&chatMsg).RecordNotFound() {
			time.Sleep(2 * time.Second)
			times = times - 1
			go doRead(msgCodes, times)
		} else {
			chatMsg.IsRead = 1
			chatMsg.UpdatedAt = time.Now()
			saveR := db.Save(&chatMsg)

			saveErr := saveR.Error
			if saveErr != nil {
				log.Printf("error: %v", saveErr)
			} else {
				if _, ok := msgCode[msgCodes]; ok {
					delete(msgCode, msgCodes)
				}
			}
		}
	}

}

func msgStore(types int, fromsccode string, tosccode string, contact string, isread int, msgcodestr string) {

	var args = ""
	args += dbuser
	args += ":"
	args += dbpassword
	args += "@("
	args += dbhost
	if dbport != "" {
		args += ":"
		args += dbport
	}
	args += ")/"
	args += dbname
	args += "?charset=utf8&parseTime=True&loc=Local"

	//db, err := gorm.Open("mysql", "root:@/eth?charset=utf8&parseTime=True&loc=Local")
	db, err := gorm.Open("mysql", args)
	if err != nil {
		log.Printf("error: %v", err)
	}
	defer db.Close()

	msgCodeInt, err := strconv.ParseInt(msgcodestr, 10, 64)
	if err != nil {
		log.Printf("error: %v", err)
	} else {
		isread = 0
		chatMsg := ChatMsg{Type: types, FromSccode: fromsccode, ToSccode: tosccode, Contact: contact, IsRead: isread, MsgCode: msgCodeInt, CreatedAt: time.Now(), UpdatedAt: time.Now()}

		created := db.Create(&chatMsg)
		createErr := created.Error
		if createErr != nil {
			log.Printf("error: %v", createErr)
		} else {
			msgCode[msgCodeInt] = true
		}
	}

}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump(hub *Hub) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:

			//fmt.Print(hub)

			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			//n := len(c.send)
			//for i := 0; i < n; i++ {
			//	w.Write(newline)
			//	w.Write(<-c.send)
			//}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	var wsUpgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	//conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	query := r.URL.Query()
	sccode := query["sccode"][0]
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), sccode: sccode}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump(hub)
	go client.readPump(hub)
}
