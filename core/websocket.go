package core

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for simplicity
	},
}

type WSMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type WSClient struct {
	conn   *websocket.Conn
	send   chan WSMessage
	userID int
}

type WSHub struct {
	clients    map[*WSClient]bool
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan WSMessage
}

var hub = &WSHub{
	clients:    make(map[*WSClient]bool),
	register:   make(chan *WSClient),
	unregister: make(chan *WSClient),
	broadcast:  make(chan WSMessage),
}

func SetupWebSocket(e *echo.Echo, db *sql.DB) {
	go hub.run()
	
	e.GET("/ws", func(c echo.Context) error {
		return handleWebSocket(c, db)
	})
}

func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("WebSocket client connected, total: %d", len(h.clients))
			
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("WebSocket client disconnected, total: %d", len(h.clients))
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

func handleWebSocket(c echo.Context, db *sql.DB) error {
	// Check authentication
	session := c.Get("session")
	if session == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
	}
	
	sessionData := session.(*Session)
	
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	
	client := &WSClient{
		conn:   conn,
		send:   make(chan WSMessage, 256),
		userID: sessionData.UserID,
	}
	
	hub.register <- client
	
	go client.writePump()
	go client.readPump()
	
	return nil
}

func (c *WSClient) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()
	
	for {
		var msg WSMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Handle incoming messages
		handleWSMessage(c, msg)
	}
}

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

func handleWSMessage(client *WSClient, msg WSMessage) {
	switch msg.Type {
	case "ping":
		client.send <- WSMessage{Type: "pong", Data: nil}
	case "get_interfaces":
		// Get real network interfaces and send update
		interfaces := getRealNetworkInterfaces()
		client.send <- WSMessage{Type: "interfaces_update", Data: interfaces}
	case "get_system_info":
		// Get real system information
		systemInfo := getRealSystemInfo()
		client.send <- WSMessage{Type: "system_info_update", Data: systemInfo}
	}
}

func BroadcastMessage(msgType string, data interface{}) {
	hub.broadcast <- WSMessage{
		Type: msgType,
		Data: data,
	}
}

func getRealNetworkInterfaces() interface{} {
	// This will be implemented to get actual network interfaces
	// For now, return empty structure to avoid mock data
	return map[string]interface{}{
		"interfaces": []interface{}{},
		"timestamp": "now",
	}
}

func getRealSystemInfo() interface{} {
	// This will be implemented to get actual system information
	// For now, return empty structure to avoid mock data
	return map[string]interface{}{
		"cpu": map[string]interface{}{},
		"memory": map[string]interface{}{},
		"storage": map[string]interface{}{},
		"network": map[string]interface{}{},
		"timestamp": "now",
	}
}
