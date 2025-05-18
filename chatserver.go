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
	upgrade     = websocket.Upgrader{}
)

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

func server() {
	http.HandleFunc("/ws", wsHandler)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func client() {
	server()
}
