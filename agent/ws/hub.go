package ws

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	mu             sync.RWMutex
	clients        map[*websocket.Conn]chan []byte
	lastSnap       []byte
	messageHandler func([]byte)
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*websocket.Conn]chan []byte)}
}

func (h *Hub) SetSnapshot(data []byte) {
	h.mu.Lock()
	h.lastSnap = data
	h.mu.Unlock()
}

func (h *Hub) SetMessageHandler(handler func([]byte)) {
	h.mu.Lock()
	h.messageHandler = handler
	h.mu.Unlock()
}

func (h *Hub) HandleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	ch := make(chan []byte, 128)
	h.mu.Lock()
	h.clients[conn] = ch
	if h.lastSnap != nil {
		conn.WriteMessage(websocket.TextMessage, h.lastSnap)
	}
	h.mu.Unlock()

	// 写协程：串行化所有写入
	go func() {
		pingTicker := time.NewTicker(15 * time.Second)
		defer pingTicker.Stop()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					return
				}
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			case <-pingTicker.C:
				conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// 读协程：阻塞直到断开
	go func() {
		defer func() {
			h.mu.Lock()
			close(ch)
			delete(h.clients, conn)
			h.mu.Unlock()
			conn.Close()
		}()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			h.mu.RLock()
			handler := h.messageHandler
			h.mu.RUnlock()
			if handler != nil {
				go handler(msg)
			}
		}
	}()
}

func (h *Hub) Broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, ch := range h.clients {
		select {
		case ch <- msg:
		default:
			// client too slow, drop
		}
	}
}
