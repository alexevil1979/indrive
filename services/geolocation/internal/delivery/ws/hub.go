// Package ws — WebSocket stub for real-time tracking (2026)
// Full implementation: broadcast driver positions to passengers, accept driver updates
package ws

import (
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Hub — stub: accepts connections, echoes or sends "connected"
type Hub struct {
	clients   map[*websocket.Conn]struct{}
	mu        sync.RWMutex
	broadcast chan []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:   make(map[*websocket.Conn]struct{}),
		broadcast: make(chan []byte, 256),
	}
}

func (h *Hub) Run() {
	for msg := range h.broadcast {
		h.mu.RLock()
		for conn := range h.clients {
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				h.mu.RUnlock()
				h.mu.Lock()
				delete(h.clients, conn)
				conn.Close()
				h.mu.Unlock()
				h.mu.RLock()
			}
		}
		h.mu.RUnlock()
	}
}

func (h *Hub) HandleConnect(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
	slog.Info("ws client connected", "remote", r.RemoteAddr)
	return conn, nil
}

func (h *Hub) Register(conn *websocket.Conn) {
	h.mu.Lock()
	h.clients[conn] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) Unregister(conn *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, conn)
	h.mu.Unlock()
	conn.Close()
}

func (h *Hub) Broadcast(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
		slog.Warn("ws broadcast channel full")
	}
}
