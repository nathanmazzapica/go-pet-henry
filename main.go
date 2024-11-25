package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"database/sql"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/google/uuid"
)



var db *sql.DB

var upgrader = websocket.Upgrader{
	// Allows all origins (just for simplicity sake, this can and should be customized in production)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	conn *websocket.Conn
}

type ClientMessage struct {
	UserID	string	`json:"userID"`
	Action	string	`json:"action"`
	Content string	`json:"content,omitempty"` 
}

type Message struct {
	Action string	`json:"action"`
	Value int		`json:"value"`
}



var (
	clients		= make(map[*Client]bool)	// Track connected clients
	counter		int							// Shared counter
	mu			sync.Mutex					// Mutex to lock/unlock
	broadcast	= make(chan int)			// Channel to broadcast updates
)


func serveHome(w http.ResponseWriter, req *http.Request) {
	user_id_cookie, err := req.Cookie("user_id")
	var uid string
	var user User
	
	// this if statement does way too much rn
	if err != nil {
		fmt.Println("error getting user id", err)
		uid = uuid.New().String()
		fmt.Println("Generated ID:", uid)

		http.SetCookie(w, &http.Cookie{
			Name:		"user_id",
			Value:		uid,
			SameSite: http.SameSiteLaxMode,
		})
		fmt.Println("Set-cookie header added for user id:", uid)
		
		// TODO: Export into createInSQL func

		user = CreateUser(uid)

		fmt.Println(user.DisplayName)

		query := "INSERT INTO users (user_id, display_name) VALUES (?,?)"
		_, err := db.Exec(query, user.UserID, user.DisplayName)

		if err != nil {
			fmt.Println("Error adding new user to db:",err)
		}

	} else {
		// TODO: Export into retrieve from SQL func
		uid = user_id_cookie.Value
		fmt.Println("loaded uid:", uid)
		result := db.QueryRow("SELECT display_name FROM users WHERE user_id = ?", uid)
		
		var displayName string

		result.Scan(&displayName)
		user.DisplayName = displayName

	}

	data := struct {
		Title string
		User string
	}{
		Title: "Pet Henry²",
		User: user.DisplayName,
	}

	tmpl := template.Must(template.ParseFiles("templates/template.html"))

	tmpl.Execute(w, data)

}


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


		
	
		rows, rerr := db.Query("SELECT * FROM users WHERE user_id = ?", "test-3")

		if rerr != nil {
			fmt.Println("error getting rows:", rerr)
		}

		// Prints user retrieved from query on line 110
		for rows.Next() {
			user := new(User)
			rerr = rows.Scan(&user.Id, &user.Pets, &user.UserID, &user.DisplayName, &user.CreatedAt)

			if rerr != nil {
				fmt.Println("Error parsing data:", rerr)
			}

			fmt.Printf("%v | %v | %v\n", user.UserID, user.DisplayName, user.Pets)
		}


		var clientMsg ClientMessage

		if err := json.Unmarshal(msg, &clientMsg); err != nil {
			fmt.Println("Error decoding JSON:", err)
			continue
		}

		switch clientMsg.Action {
		case "pet":
			mu.Lock()
			counter++
			fmt.Printf("User %s pet henry! Total pets is now %d\n", clientMsg.UserID, counter)

			mu.Unlock()

			broadcast <- counter

			
			_, err := db.Exec("UPDATE users SET pets = pets + 1 WHERE user_id = ?", clientMsg.UserID)	

			if err != nil {
				fmt.Println("error updating pets:", err)
			}


		case "connect":
			fmt.Printf("New user connected!\n")
		default:
			fmt.Printf("Unknown action: %s from user %s\n", clientMsg.Action, clientMsg.UserID)
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
				Action: "counter",
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

	var dberr error
	db, dberr = sql.Open("sqlite3", "henry.db")
	defer db.Close()

	if dberr != nil {
		fmt.Println("Error opening database:", dberr)
	}

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
