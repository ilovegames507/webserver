package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins
	},
}

var (
	mainMu  sync.Mutex
	clients = make(map[*websocket.Conn]string)
)

// Broadcast message to all clients except the sender
func broadcastToClients(message []byte, sender *websocket.Conn) {
	mainMu.Lock()
	defer mainMu.Unlock()
	for conn := range clients {
		if conn != sender {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Broadcast error:", err)
				conn.Close()
				delete(clients, conn)
			}
		}
	}
}

func wshandler(w http.ResponseWriter, r *http.Request) {
	// Check password query parameter before upgrading
	password := r.URL.Query().Get("password")
	if password != "yourpassword" { // replace with your actual password or env var
		http.Error(w, "Unauthorized - invalid password", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	// Add new connection
	mainMu.Lock()
	clients[conn] = r.RemoteAddr
	mainMu.Unlock()

	log.Printf("Client connected: %s", r.RemoteAddr)

	// Read messages loop
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("Received: %s", msg)
		broadcastToClients(msg, conn)
	}

	// Cleanup connection
	mainMu.Lock()
	delete(clients, conn)
	mainMu.Unlock()

	log.Printf("Client disconnected: %s", r.RemoteAddr)
}

func main() {
	http.HandleFunc("/ws", wshandler)
	log.Println("Server started on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
