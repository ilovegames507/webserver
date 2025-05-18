package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type user struct {
	userid   string
	username string
}

func webpassword() string {
	return os.Getenv("WEBSOCKET_PASSWORD")
}

func checkPassword(password string) error {
	if password != webpassword() {
		return errors.New("invalid password")
	}
	return nil
}

var (
	connections = make(map[*websocket.Conn]string)
	mu          sync.Mutex
)

var upgrader1 = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all origins (adjust as needed)
	},
}

func broadcast(message []byte, sender *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()
	for conn := range connections {
		if conn != sender {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Broadcast error:", err)
				conn.Close()
				delete(connections, conn)
			}
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	password := r.URL.Query().Get("password")
	if err := checkPassword(password); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println("Unauthorized connection attempt from", r.RemoteAddr)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	connections[conn] = r.RemoteAddr
	mu.Unlock()

	log.Printf("Client connected: %s", r.RemoteAddr)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}
		log.Printf("Received: %s", msg)
		broadcast(msg, conn)
	}

	mu.Lock()
	delete(connections, conn)
	mu.Unlock()
	log.Printf("Client disconnected: %s", r.RemoteAddr)
}

func RunServer() {
	http.HandleFunc("/ws", wsHandler)
	log.Println("Server started on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
