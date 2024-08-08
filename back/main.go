package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type User struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	JoinedAt time.Time `json:"joinedAt"`
	IsHost   bool      `json:"isHost"`
}

type Message struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	users    = make(map[*websocket.Conn]User)
	usersMux sync.Mutex
	host     *websocket.Conn
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			removeUser(conn)
			break
		}

		var message Message
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Println(err)
			continue
		}

		switch message.Type {
		case "connect":
			addUser(conn, User{ID: message.ID, Name: message.Name, JoinedAt: time.Now()})
		}
	}
}

func addUser(conn *websocket.Conn, user User) {
	usersMux.Lock()
	defer usersMux.Unlock()

	if host == nil {
		host = conn
		user.IsHost = true
	}
	users[conn] = user
	broadcastUsers()
}

func removeUser(conn *websocket.Conn) {
	usersMux.Lock()
	defer usersMux.Unlock()

	isHost := users[conn].IsHost
	delete(users, conn)

	if isHost {
		selectNewHost()
	}
	broadcastUsers()
}

func selectNewHost() {
	if len(users) == 0 {
		host = nil
		return
	}

	var oldestUser *websocket.Conn
	var oldestTime time.Time

	for conn, user := range users {
		if oldestUser == nil || user.JoinedAt.Before(oldestTime) {
			oldestUser = conn
			oldestTime = user.JoinedAt
		}
	}

	host = oldestUser
	user := users[host]
	user.IsHost = true
	users[host] = user
}

func broadcastUsers() {
	usersList := make([]User, 0, len(users))
	for _, user := range users {
		usersList = append(usersList, user)
	}

	sort.Slice(usersList, func(i, j int) bool {
		return usersList[i].JoinedAt.Before(usersList[j].JoinedAt)
	})

	message, _ := json.Marshal(map[string]interface{}{
		"type":  "users",
		"users": usersList,
	})

	for conn := range users {
		conn.WriteMessage(websocket.TextMessage, message)
	}
}

func sendToHost(message []byte) {
	if host != nil {
		host.WriteMessage(websocket.TextMessage, message)
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("Start")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
