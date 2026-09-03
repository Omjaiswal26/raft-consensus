package api

import (
	"raft-consensus/internal/raft"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	send chan any
	done chan struct{}
}

type Hub struct {
	mu      sync.Mutex
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*Client]struct{})}
}

func (h *Hub) Register(conn *websocket.Conn) *Client {
	c := &Client{
		send: make(chan any, 256),
		done: make(chan struct{}),
	}
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()

	go h.writePump(conn, c)
	return c
}

func (h *Hub) writePump(conn *websocket.Conn, c *Client) {
	defer h.Unregister(c)
	defer conn.Close()

	for msg := range c.send {
		_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		if err := conn.WriteJSON(msg); err != nil {
			return
		}
	}
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; !ok {
		h.mu.Unlock()
		return
	}
	delete(h.clients, c)
	h.mu.Unlock()

	close(c.done)
	close(c.send)
}

func (h *Hub) Send(c *Client, msg any) bool {
	select {
	case <-c.done:
		return false
	default:
	}

	select {
	case c.send <- msg:
		return true
	case <-c.done:
		return false
	}
}

func (h *Hub) Emit(e raft.Event) {
	h.mu.Lock()
	clients := make([]*Client, 0, len(h.clients))
	for c := range h.clients {
		clients = append(clients, c)
	}
	h.mu.Unlock()

	for _, c := range clients {
		h.Send(c, e)
	}
}
