package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/netmap/netmap/internal/core/models"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Client struct {
	hub       *Hub
	conn      *websocket.Conn
	send      chan []byte
	mu        sync.Mutex
	closeOnce sync.Once
}

func (c *Client) close() {
	c.closeOnce.Do(func() { c.conn.Close() })
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	stop       chan struct{}
	done       chan struct{}
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte, 256),
		stop:       make(chan struct{}),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Register(c *Client) {
	h.register <- c
}

func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

func (h *Hub) Broadcast(event models.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("ws broadcast marshal error: %v", err)
		return
	}
	select {
	case h.broadcast <- data:
	default:
		log.Println("ws broadcast channel full, dropping event")
	}
}

func (h *Hub) Run() {
	defer close(h.done)
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					go func(c *Client) {
						select {
						case h.unregister <- c:
						case <-h.stop:
						}
					}(client)
				}
			}
			h.mu.RUnlock()
		case <-h.stop:
			return
		}
	}
}

func (h *Hub) Stop() {
	close(h.stop)
	<-h.done
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}
	h.register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	defer c.close()
	for msg := range c.send {
		c.mu.Lock()
		err := c.conn.WriteMessage(websocket.TextMessage, msg)
		c.mu.Unlock()
		if err != nil {
			return
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		select {
		case c.hub.unregister <- c:
		case <-c.hub.stop:
		}
		c.close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
