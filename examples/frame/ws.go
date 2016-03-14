package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type wsMessage struct {
	User    string `json:"user"`
	Channel string `json:"channel"`
	Text    string `json:"text"`
	TS      int64  `json:"ts"`
}

type wsMessages []*wsMessage

func (m wsMessages) Len() int {
	return len(m)
}

func (m wsMessages) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m wsMessages) Less(i, j int) bool {
	return m[i].TS < m[j].TS
}

type wsHandler struct {
	clients  []*websocket.Conn // The connections we need to update
	messages chan *wsMessage   // The messages we should send to our clients
	mux      sync.Mutex        // protect the clients from concurrent changes
}

func newWSHandler() *wsHandler {
	return &wsHandler{messages: make(chan *wsMessage)}
}

func (ws *wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if *debug {
		log.Printf("Upgrading WS from %v", r.RemoteAddr)
	}
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Fail to establish WebSocket conn , %v", err)
		panic(err)
	}
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	ws.mux.Lock()
	defer ws.mux.Unlock()
	ws.clients = append(ws.clients, conn)
	go ws.readLoop(conn)
}

func (ws *wsHandler) send(msg *wsMessage) {
	ws.messages <- msg
}

func (ws *wsHandler) stop() {
	close(ws.messages)
}

// readLoop reads messages from the websocket connection and ignores them
func (ws *wsHandler) readLoop(conn *websocket.Conn) {
	defer conn.Close()
	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
	if *debug {
		log.Printf("Got error from read loop for client [%v] - closing\n", conn.RemoteAddr())
	}
	ws.mux.Lock()
	defer ws.mux.Unlock()
	client := -1
	for i, c := range ws.clients {
		if c == conn {
			client = i
			break
		}
	}
	if client >= 0 {
		ws.clients = append(ws.clients[:client], ws.clients[client+1:]...)
	}
}

// writeLoop will clear the messages from the channel and send them to all clients
func (ws *wsHandler) writeLoop() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ws.mux.Lock()
		defer ws.mux.Unlock()
		for _, c := range ws.clients {
			c.Close()
		}
		ws.clients = []*websocket.Conn{}
		ticker.Stop()
	}()
	for {
		select {
		case message, ok := <-ws.messages:
			if !ok {
				return
			}
			ws.mux.Lock()
			for _, c := range ws.clients {
				c.WriteJSON(message)
			}
			ws.mux.Unlock()
		case <-ticker.C:
			var clients []*websocket.Conn // the new clients filtered from the bad ones
			ws.mux.Lock()
			for _, c := range ws.clients {
				if err := c.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					log.Printf("Error writing ping to client %s - %v\n", c.RemoteAddr().String(), err)
					c.Close()
				} else {
					clients = append(clients, c)
				}
			}
			if len(clients) != len(ws.clients) {
				ws.clients = clients
			}
			ws.mux.Unlock()
		}
	}
}
