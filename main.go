package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
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

type ChangeDisplayNameRequest struct {
	DisplayName string `json:"displayName"`
}

type ChangeDisplayNameResponse struct{
	Success bool	`json:"success"`
	Message string	`json:"message"`
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


func validateName(displyName string) error {
	if len(displyName) < 4 || len(displyName) > 25 {
		return errors.New("Name must be between 4 and 25 characters.")
	}

	if strings.ContainsRune(displyName, ' ') {
		return errors.New("Name cannot contain spaces!")
	}

	if strings.ContainsAny(displyName, "\"\b\n\r\t%!@#$%^&*()-+=;:/.,<>`~_'") {
		return errors.New("Name may only contain letters and numbers")
	}

	return nil
}

func changeDisplayName(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed!", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)

	if err != nil {
		http.Error(w, "Unabled to read request body", http.StatusBadRequest)
	}
	
	var changeRequest ChangeDisplayNameRequest
	err = json.Unmarshal(body, &changeRequest)

	if err != nil {
		fmt.Println("Error parsing display name change request", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
	}

	fmt.Println(body)

	responseSuccess := true
	responseMessage := "DisplayName changed to " + changeRequest.DisplayName


	err = validateName(changeRequest.DisplayName)
	if err != nil {
		fmt.Println("Invalid name:", err)
		responseSuccess = false
		responseMessage = err.Error()
	}

	user_id_cookie, err := req.Cookie("user_id")

	user_id_string := user_id_cookie.Value

	result, err := db.Exec("UPDATE users SET display_name = ? WHERE user_id = ?",
							changeRequest.DisplayName, user_id_string)
	
	affectedRows, _ := result.RowsAffected()
	fmt.Println("Affected", affectedRows, "rows")
	

	response := ChangeDisplayNameResponse{
		Success: responseSuccess,
		Message: responseMessage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

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

		var pets int
		result = db.QueryRow("SELECT pets FROM users WHERE user_id = ?", uid)

		result.Scan(&pets)
		user.Pets = pets



	}

	data := struct {
		Title string
		User string
		UserPets int
		TotalPets int
	}{
		Title: "Pet Henry²",
		User: user.DisplayName,
		UserPets: user.Pets,
		TotalPets: counter,
	}

	tmpl := template.Must(template.ParseFiles("templates/template.html"))

	tmpl.Execute(w, data)

}

func validateUserFromCookies(r *http.Request) (string, error) {
	cookie, err := r.Cookie("user_id")

	if err != nil {
		return "", err
	}

	userID := cookie.Value
	return userID, nil
}


func handleConnections(w http.ResponseWriter, r *http.Request) {

	userID, err := validateUserFromCookies(r)

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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
		fmt.Println(userID)
		_, msg, err := conn.ReadMessage()

		if err != nil {
			fmt.Println("Error reading message:",err)
			break
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

			
			_, err := db.Exec("UPDATE users SET pets = pets + 1 WHERE user_id = ?", userID)	

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

	result := db.QueryRow("SELECT SUM(pets) FROM users")

	result.Scan(&counter)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))


	http.HandleFunc("/",serveHome)
	
	http.HandleFunc("/ws", handleConnections)

	http.HandleFunc("/cd", changeDisplayName)

	go handleBroadcasts()

	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		fmt.Println("something went horribly wrong", err)
	}

}
