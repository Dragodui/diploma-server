package ws

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Dragodui/diploma-server/internal/config"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const (
	pingInterval = 30 * time.Second
	pongTimeout  = 60 * time.Second
)

// handler for all ws connections
type WSHandler struct {
	Upgrader websocket.Upgrader
	Clients  map[*websocket.Conn]bool
	Mu       sync.Mutex
}

func NewWSHandler(cfg *config.Config) *WSHandler {
	return &WSHandler{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				if cfg.Mode == "dev" {
					return true
				}

				// mobile clients (e.g. Expo Go) may not send an Origin header
				origin := r.Header.Get("Origin")
				if origin == "" {
					return true
				}

				if origin == cfg.ClientURL || origin == "http://"+cfg.ClientURL || origin == "https://"+cfg.ClientURL {
					return true
				}

				return false
			},
		},
		Clients: make(map[*websocket.Conn]bool),
	}
}

func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request, cache *redis.Client) {
	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// lock to update connections list
	h.Mu.Lock()
	h.Clients[conn] = true
	h.Mu.Unlock()

	go h.readPump(conn)
	go h.subscribeToCache(conn, cache)
}

// readPump reads from the connection to handle pong responses and detect disconnects.
func (h *WSHandler) readPump(conn *websocket.Conn) {
	defer h.removeClient(conn)

	conn.SetReadDeadline(time.Now().Add(pongTimeout))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongTimeout))
		return nil
	})

	for {
		// ReadMessage blocks until a message arrives or the connection errors out.
		// We discard client messages; this loop exists only to detect disconnects.
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *WSHandler) subscribeToCache(conn *websocket.Conn, cache *redis.Client) {
	defer h.removeClient(conn)

	pubsub := cache.Subscribe(context.Background(), "updates")
	defer pubsub.Close()

	// start sending pings to keep the connection alive
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	msgCh := pubsub.Channel()

	for {
		select {
		case msg, ok := <-msgCh:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				log.Printf("Error writing WS message: %v", err)
				return
			}
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
				log.Printf("Ping failed: %v", err)
				return
			}
		}
	}
}

// removeClient safely closes the connection and removes it from the client map.
func (h *WSHandler) removeClient(conn *websocket.Conn) {
	h.Mu.Lock()
	if _, ok := h.Clients[conn]; ok {
		delete(h.Clients, conn)
		conn.Close()
	}
	h.Mu.Unlock()
}
