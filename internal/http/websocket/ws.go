package ws

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// handler for all ws connections
type WSHandler struct {
	Upgrader websocket.Upgrader
	Clients  map[*websocket.Conn]bool
	Mu       sync.Mutex
}

func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request, cache *redis.Client) {
	conn, _ := h.Upgrader.Upgrade(w, r, nil)

	// lock to update connections list
	h.Mu.Lock()
	h.Clients[conn] = true
	h.Mu.Unlock()

	go h.subscribeToCache(conn, cache)
}

func (h *WSHandler) subscribeToCache(conn *websocket.Conn, cache *redis.Client) {
	// subscribe to redis pubsub
	pubsub := cache.Subscribe(context.Background(), "updates")
	defer pubsub.Close()

	// handle messages
	for msg := range pubsub.Channel() {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			log.Printf("Error while writing messages: %s\n", err.Error())
			break
		}
	}
}
