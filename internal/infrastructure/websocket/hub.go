package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// クライアントへの書き込み待ち時間
	writeWait = 10 * time.Second
	// pong メッセージ受信までの待ち時間
	pongWait = 60 * time.Second
	// ping メッセージ送信間隔（pongWait より短くする）
	pingPeriod = (pongWait * 9) / 10
	// 最大メッセージサイズ
	maxMessageSize = 512
)

// Message は WebSocket で送受信するメッセージ
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// Client は WebSocket クライアント接続を表す
type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	userID         string
	send           chan []byte
	subscribedJobs map[string]bool
	mu             sync.Mutex
}

// Hub は WebSocket 接続を管理する中央ハブ
type Hub struct {
	// ユーザー ID ごとのクライアント
	clients map[string]map[*Client]bool
	// 登録リクエスト
	register chan *Client
	// 登録解除リクエスト
	unregister chan *Client
	// mutex
	mu sync.RWMutex
}

// NewHub は新しい Hub を作成する
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run は Hub のメインループを起動する
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.userID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// SendToUser はユーザーにメッセージを送信する
func (h *Hub) SendToUser(userID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for client := range clients {
		select {
		case client.send <- data:
		default:
			// バッファがいっぱいの場合はスキップ
		}
	}
}

// SendToJob はジョブを購読しているクライアントにメッセージを送信する
func (h *Hub) SendToJob(jobID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, clients := range h.clients {
		for client := range clients {
			client.mu.Lock()
			if client.subscribedJobs[jobID] {
				select {
				case client.send <- data:
				default:
				}
			}
			client.mu.Unlock()
		}
	}
}

// RegisterClient は新しいクライアントを登録する
func (h *Hub) RegisterClient(conn *websocket.Conn, userID string) *Client {
	client := &Client{
		hub:            h,
		conn:           conn,
		userID:         userID,
		send:           make(chan []byte, 256),
		subscribedJobs: make(map[string]bool),
	}
	h.register <- client
	return client
}

// ReadPump はクライアントからのメッセージを読み取るループ
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			// 予期しない切断はログで処理（呼び出し側）
			break
		}

		c.handleMessage(message)
	}
}

// WritePump はクライアントへのメッセージ送信ループ
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// キュー内のメッセージをまとめて送信
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage はクライアントからのメッセージを処理する
func (c *Client) handleMessage(data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	switch msg.Type {
	case "subscribe":
		c.handleSubscribe(msg.Payload)
	case "unsubscribe":
		c.handleUnsubscribe(msg.Payload)
	case "ping":
		c.sendPong()
	}
}

// handleSubscribe はジョブ購読リクエストを処理する
func (c *Client) handleSubscribe(payload interface{}) {
	p, ok := payload.(map[string]interface{})
	if !ok {
		return
	}
	jobID, ok := p["jobId"].(string)
	if !ok {
		return
	}

	c.mu.Lock()
	c.subscribedJobs[jobID] = true
	c.mu.Unlock()
}

// handleUnsubscribe はジョブ購読解除リクエストを処理する
func (c *Client) handleUnsubscribe(payload interface{}) {
	p, ok := payload.(map[string]interface{})
	if !ok {
		return
	}
	jobID, ok := p["jobId"].(string)
	if !ok {
		return
	}

	c.mu.Lock()
	delete(c.subscribedJobs, jobID)
	c.mu.Unlock()
}

// sendPong は pong レスポンスを送信する
func (c *Client) sendPong() {
	msg := Message{Type: "pong"}
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	select {
	case c.send <- data:
	default:
	}
}
