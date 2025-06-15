package core

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	upgrader   websocket.Upgrader
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	manager := &WebSocketManager{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow connections from any origin
			},
		},
	}
	
	go manager.run()
	return manager
}

// run handles WebSocket connections
func (manager *WebSocketManager) run() {
	for {
		select {
		case client := <-manager.register:
			manager.clients[client] = true
			log.Println("WebSocket client connected")
			
		case client := <-manager.unregister:
			if _, ok := manager.clients[client]; ok {
				delete(manager.clients, client)
				client.Close()
				log.Println("WebSocket client disconnected")
			}
			
		case message := <-manager.broadcast:
			for client := range manager.clients {
				if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("WebSocket write error: %v", err)
					delete(manager.clients, client)
					client.Close()
				}
			}
		}
	}
}

// HandleWebSocket handles WebSocket connections
func (manager *WebSocketManager) HandleWebSocket(c echo.Context) error {
	conn, err := manager.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	
	manager.register <- conn
	
	// Handle messages from client
	go func() {
		defer func() {
			manager.unregister <- conn
		}()
		
		for {
			var msg WebSocketMessage
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}
			
			// Handle different message types
			manager.handleMessage(msg)
		}
	}()
	
	return nil
}

// handleMessage handles incoming WebSocket messages
func (manager *WebSocketManager) handleMessage(msg WebSocketMessage) {
	switch msg.Type {
	case "ping":
		manager.SendMessage("pong", map[string]string{"status": "ok"})
	case "get_interfaces":
		// Handle interface status request
		manager.SendMessage("interfaces_update", map[string]string{"status": "updated"})
	}
}

// SendMessage broadcasts a message to all connected clients
func (manager *WebSocketManager) SendMessage(msgType string, data interface{}) {
	msg := WebSocketMessage{
		Type: msgType,
		Data: data,
	}
	
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}
	
	select {
	case manager.broadcast <- jsonData:
	default:
		log.Println("WebSocket broadcast channel is full")
	}
}
