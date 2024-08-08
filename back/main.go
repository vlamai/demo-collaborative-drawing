package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsHost   bool   `json:"isHost"`
	Conn     *websocket.Conn `json:"-"`
}

type Message struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
	Size CanvasSize `json:"size,omitempty"`
}

type CanvasSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

var (
	users    = make(map[string]User)
	usersMux sync.Mutex
	host     *User
	canvasSize CanvasSize
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
		var message Message
		err := conn.ReadJSON(&message)
		if err != nil {
			log.Println(err)
			removeUser(message.ID)
			break
		}

		switch message.Type {
		case "connect":
			addUser(message.ID, message.Name, conn)
		case "canvasSize":
			updateCanvasSize(message.Size)
		}
	}
}

func addUser(id, name string, conn *websocket.Conn) {
	usersMux.Lock()
	defer usersMux.Unlock()

	newUser := User{ID: id, Name: name, Conn: conn}
	if host == nil {
		host = &newUser
		newUser.IsHost = true
	}
	users[id] = newUser

	broadcastUsers()
	if !newUser.IsHost {
		sendCanvasSizeToUser(conn)
	}
}

func removeUser(id string) {
	usersMux.Lock()
	defer usersMux.Unlock()

	if users[id].IsHost {
		host = nil
		for _, user := range users {
			if user.ID != id {
				host = &user
				user.IsHost = true
				break
			}
		}
	}
	delete(users, id)
	broadcastUsers()
}

func updateCanvasSize(size CanvasSize) {
	canvasSize = size
	broadcastCanvasSize()
}

func broadcastUsers() {
	usersList := make([]User, 0, len(users))
	for _, user := range users {
		usersList = append(usersList, User{ID: user.ID, Name: user.Name, IsHost: user.IsHost})
	}

	message, _ := json.Marshal(map[string]interface{}{
		"type":  "users",
		"users": usersList,
	})

	for _, user := range users {
		user.Conn.WriteMessage(websocket.TextMessage, message)
	}
}

func broadcastCanvasSize() {
	message, _ := json.Marshal(map[string]interface{}{
		"type": "canvasSize",
		"size": canvasSize,
	})

	for _, user := range users {
		if !user.IsHost {
			user.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func sendCanvasSizeToUser(conn *websocket.Conn) {
	message, _ := json.Marshal(map[string]interface{}{
		"type": "canvasSize",
		"size": canvasSize,
	})
	conn.WriteMessage(websocket.TextMessage, message)
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
