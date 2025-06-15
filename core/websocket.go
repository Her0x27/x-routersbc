package core

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin in development
	},
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// WSClient represents a WebSocket client
type WSClient struct {
	conn   *websocket.Conn
	send   chan WSMessage
	hub    *WSHub
	userID int
}

// WSHub maintains active clients and broadcasts messages
type WSHub struct {
	clients    map[*WSClient]bool
	broadcast  chan WSMessage
	register   chan *WSClient
	unregister chan *WSClient
}

// NewWSHub creates a new WebSocket hub
func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[*WSClient]bool),
		broadcast:  make(chan WSMessage),
		register:   make(chan *WSClient),
		unregister: make(chan *WSClient),
	}
}

// Run starts the WebSocket hub
func (h *WSHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("WebSocket client connected. Total: %d", len(h.clients))

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("WebSocket client disconnected. Total: %d", len(h.clients))
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastMessage sends a message to all connected clients
func (h *WSHub) BroadcastMessage(msgType string, data interface{}) {
	message := WSMessage{
		Type: msgType,
		Data: data,
	}
	h.broadcast <- message
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(hub *WSHub) echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return err
		}

		// Get user session
		session := c.Get("session").(*Session)
		
		client := &WSClient{
			conn:   conn,
			send:   make(chan WSMessage, 256),
			hub:    hub,
			userID: session.UserID,
		}

		client.hub.register <- client

		// Start goroutines for reading and writing
		go client.writePump()
		go client.readPump()

		return nil
	}
}

// readPump handles reading messages from the WebSocket connection
func (c *WSClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var message WSMessage
		err := c.conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages
		c.handleMessage(message)
	}
}

// writePump handles writing messages to the WebSocket connection
func (c *WSClient) writePump() {
	defer c.conn.Close()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (c *WSClient) handleMessage(message WSMessage) {
	switch message.Type {
	case "ping":
		c.send <- WSMessage{Type: "pong", Data: nil}
	case "get_interfaces":
		// Get current network interfaces and send to client
		// This would integrate with the network service
		c.send <- WSMessage{
			Type: "interfaces_update",
			Data: map[string]interface{}{
				"timestamp": "current_time",
				"interfaces": []interface{}{},
			},
		}
	default:
		log.Printf("Unknown WebSocket message type: %s", message.Type)
	}
}

// Global WebSocket hub instance
var GlobalWSHub *WSHub

// InitWebSocket initializes the global WebSocket hub
func InitWebSocket() {
	GlobalWSHub = NewWSHub()
	go GlobalWSHub.Run()
}
