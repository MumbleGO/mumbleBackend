package server

import (
	"encoding/json"
	"fmt"
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

type Message struct {
	Type    string   `json:"type"`
	Content []string `json:"content"`
}

type NewMessage struct {
	Id        string    `json:"id"`
	Body      string    `json:"body"`
	SenderId  string    `json:"senderId"`
	CreatedAt time.Time `json:"createdAt"`
}

var userSocketMap = struct {
	sync.RWMutex
	connections map[string]*websocket.Conn
}{connections: make(map[string]*websocket.Conn)}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
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

	message := Message{
		Type:    "getOnlineUsers",
		Content: onlineUsers,
	}
	fmt.Println(message)

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

func handleSendMessageWS(w http.ResponseWriter, r *http.Request, sender string, receiver string) {
	senderId := sender
	receiverId := receiver

	if senderId == "" || receiverId == "" {
		http.Error(w, "Missing senderId or receiverId", http.StatusBadRequest)
		return
	}

	var newMessage struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&newMessage); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	userSocketMap.RLock()
	receiverConn, exists := userSocketMap.connections[receiverId]
	userSocketMap.RUnlock()

	if exists {
		message := NewMessage{
			Id:        receiverId,
			SenderId:  senderId,
			Body:      newMessage.Content,
			CreatedAt: time.Now(),
		}

		err := receiverConn.WriteJSON(message)
		if err != nil {
			log.Printf("Error sending message to %s: %v", receiverId, err)
			http.Error(w, "Error sending message", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(message)
	} else {
		http.Error(w, "User not connected", http.StatusNotFound)
	}
}
