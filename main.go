package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)


func serveHome(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)


	data := struct {
		Title string
		User string
	}{
		Title: "Welcome Page",
		User: "Nathan",
	}

	tmpl := template.Must(template.ParseFiles("templates/template.html"))

	tmpl.Execute(w, data)

}

var upgrader = websocket.Upgrader{
	// Allows all origins (just for simplicity sake, this can and should be customized in production)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn *websocket.Conn
}
// TODO: Replace `Message` struct with ClientMessage
type ClientMessage struct {
	UserID	string	`json:"userID"`
	Action	string	`json:"action"`
	Content string	`json:"content"` 
}

type Message struct {
	Type string `json:"type"`
	Value int	`json:"value"`
}


var (
	clients		= make(map[*Client]bool)	// Track connected clients
	counter		int							// Shared counter
	mu			sync.Mutex					// Mutex to lock/unlock
	broadcast	= make(chan int)			// Channel to broadcast updates
)


func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		fmt.Println("error upgrading connection:", err)
		return
	}
	client := &Client{conn: conn}

	defer func() {
		mu.Lock()
		delete(clients, client)
		mu.Unlock()
		conn.Close()
	}()

	// Register new client
	mu.Lock()
	clients[client] = true
	mu.Unlock()

	fmt.Println("Client connected")

	for {
		_, msg, err := conn.ReadMessage()

		if err != nil {
			fmt.Println("Error reading message:",err)
			break
		}

		if string(msg) == "pet" {
			mu.Lock()
			counter++
			fmt.Println("Counter incremented:", counter)
			mu.Unlock()

			broadcast <- counter
		}


		fmt.Printf("Received: %s\n", msg)
	}
}


func handleBroadcasts() {
	for {
		// Receives updates on broadcast Channel
		updatedCounter := <-broadcast

		// Send update to clients
		mu.Lock()

		for client := range clients {

			msg := Message{
				Type: "counter",
				Value: updatedCounter,
			}

			jsonData, err := json.Marshal(msg)

			if err != nil {
				fmt.Println("Error serializing JSON:", err)
				return
			}

			err = client.conn.WriteMessage(websocket.TextMessage, jsonData)

			if err != nil {
				fmt.Println("Error broadcasting message:", err)
				client.conn.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}


func main() {
	fmt.Println("Hello, Go!")


	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))


	http.HandleFunc("/",serveHome)
	
	http.HandleFunc("/ws", handleConnections)

	go handleBroadcasts()

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("something went horribly wrong", err)
	}

}
