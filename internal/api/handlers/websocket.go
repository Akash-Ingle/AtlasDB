package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/pkg/models"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// EventBroadcaster fans out events to all connected WebSocket clients.
type EventBroadcaster struct {
	mu      sync.RWMutex
	clients map[*wsClient]struct{}
	logger  zerolog.Logger
}

type wsClient struct {
	conn   *websocket.Conn
	send   chan []byte
	done   chan struct{}
}

func NewEventBroadcaster(logger zerolog.Logger) *EventBroadcaster {
	return &EventBroadcaster{
		clients: make(map[*wsClient]struct{}),
		logger:  logger,
	}
}

// Broadcast sends events to all connected clients.
// Called by the stream processor after storing events.
func (b *EventBroadcaster) Broadcast(events []models.Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.clients) == 0 {
		return
	}

	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			continue
		}

		for client := range b.clients {
			select {
			case client.send <- data:
			default:
				// Client too slow, drop the message
			}
		}
	}
}

func (b *EventBroadcaster) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

func (b *EventBroadcaster) register(c *wsClient) {
	b.mu.Lock()
	b.clients[c] = struct{}{}
	b.mu.Unlock()
	b.logger.Debug().Int("clients", len(b.clients)).Msg("WebSocket client connected")
}

func (b *EventBroadcaster) unregister(c *wsClient) {
	b.mu.Lock()
	delete(b.clients, c)
	b.mu.Unlock()
	b.logger.Debug().Int("clients", len(b.clients)).Msg("WebSocket client disconnected")
}

type WebSocketHandler struct {
	broadcaster *EventBroadcaster
	logger      zerolog.Logger
}

func NewWebSocketHandler(broadcaster *EventBroadcaster, logger zerolog.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		broadcaster: broadcaster,
		logger:      logger,
	}
}

// GET /api/v1/events/stream (WebSocket upgrade)
func (h *WebSocketHandler) Stream(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error().Err(err).Msg("WebSocket upgrade failed")
		return
	}

	client := &wsClient{
		conn: conn,
		send: make(chan []byte, 256),
		done: make(chan struct{}),
	}

	h.broadcaster.register(client)

	go h.writePump(client)
	go h.readPump(client)
}

func (h *WebSocketHandler) writePump(c *wsClient) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		h.broadcaster.unregister(c)
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.done:
			return
		}
	}
}

func (h *WebSocketHandler) readPump(c *wsClient) {
	defer func() {
		close(c.done)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
	}
}

// StartBroadcastFromRedis subscribes to Redis pub/sub and broadcasts to WebSocket clients.
// This allows the processor (separate process) to notify the API server of new events.
func (b *EventBroadcaster) StartBroadcastFromRedis(ctx context.Context, redisAddr string) {
	// In the current architecture, the broadcaster is called directly by the processor
	// when running in the same process, or via Redis pub/sub when running separately.
	// For Phase 1, direct calls are sufficient.
	<-ctx.Done()
}
