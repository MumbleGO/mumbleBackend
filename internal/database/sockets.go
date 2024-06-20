package database

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "http://localhost:5173" || origin == "https://your-production-url" {
			return true
		}
		return false
	},
}

type Messagews struct {
	Type    string   `json:"type"`
	Content []string `json:"content"`
}

type NewMessage struct {
	Id          string    `json:"id"`
	Body        string    `json:"body"`
	SenderId    string    `json:"senderId"`
	CreatedAt   time.Time `json:"createdAt"`
	ShouldShake bool      `json:"shouldShake,omitempty"`
}

var userSocketMap = struct {
	sync.RWMutex
	connections map[string]*websocket.Conn
}{connections: make(map[string]*websocket.Conn)}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	userId := r.URL.Query().Get("userId")
	if userId == "" {
		log.Println("Missing userId in query parameters")
		return
	}

	userSocketMap.Lock()
	userSocketMap.connections[userId] = conn
	userSocketMap.Unlock()

	log.Printf("User connected: %s", userId)

	broadcastOnlineUsers()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Read error for %s: %v", userId, err)
			break
		}
	}

	userSocketMap.Lock()
	delete(userSocketMap.connections, userId)
	userSocketMap.Unlock()
	log.Printf("User disconnected: %s", userId)

	broadcastOnlineUsers()
}

func broadcastOnlineUsers() {
	userSocketMap.RLock()
	onlineUsers := make([]string, 0, len(userSocketMap.connections))
	for userId := range userSocketMap.connections {
		onlineUsers = append(onlineUsers, userId)
	}
	userSocketMap.RUnlock()

	message := Messagews{
		Type:    "getOnlineUsers",
		Content: onlineUsers,
	}

	for userId, conn := range userSocketMap.connections {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Error sending online users to %s: %v", userId, err)
			conn.Close()
			userSocketMap.RUnlock()
			userSocketMap.Lock()
			delete(userSocketMap.connections, userId)
			userSocketMap.Unlock()
			userSocketMap.RLock()
		}
	}
}

func notifyReceiver(receiverId string, newMessage Message) {
	userSocketMap.RLock()
	defer userSocketMap.RUnlock()

	if conn, ok := userSocketMap.connections[receiverId]; ok {
		message := NewMessage{
			Id:        newMessage.ID,
			Body:      newMessage.Body,
			SenderId:  newMessage.SenderID,
			CreatedAt: newMessage.CreatedAt,
		}
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Error sending message to receiver %s: %v", receiverId, err)
		}
	}
}
