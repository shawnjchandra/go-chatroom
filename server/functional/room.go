package server_module

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

/*
	Khusus Attribute channel pada struct mengambil inspirasi dari :
	- claude.ai
*/

var (
	mu       sync.Mutex
	roomMenu = "[In-Room Menu]\n" +
		"1. /chat <message>\n" +
		"2. /list\n" +
		"3. /leave\n"
)

type Room struct {
	Room_id   int
	Room_name string
	Users     map[string]User

	Join      chan User
	Leave     chan User
	Broadcast chan Message
}

func CreateRoom(room_name string, id int) Room {
	// id := len(rooms)
	return Room{
		Room_id:   id,
		Room_name: room_name,
		Users:     make(map[string]User),
		Join:      make(chan User, 99),
		Leave:     make(chan User, 99),
		Broadcast: make(chan Message, 99),
	}
}

func (r *Room) Run() {
	for {

		select {
		case u := <-r.Join:
			r.JoinRoom(u)
			u.SendNotification(roomMenu)
		case u := <-r.Leave:
			r.LeaveRoom(u)

		case message := <-r.Broadcast:

			r.handleChat(message)

		}
	}
}

func (r *Room) JoinRoom(user User) error {
	// room := rooms[room_id]

	if _, ok := r.Users[user.Name]; !ok {
		r.Users[user.Name] = user

		// count := user.Counter
		// user.ListRoom[count] = r.Room_name
		// r.Join <- user

		r.NotificationToAll(fmt.Sprintf("%s has joined the room %s", user.Name, r.Room_name))

		user.Counter++
		return nil
	}

	return errors.New("can't join room")

}

func (r *Room) LeaveRoom(user User) {
	r.NotificationToAll(fmt.Sprintf("%s has left the room %s", user.Name, r.Room_name))

	mu.Lock()

	delete(r.Users, user.Name)

	mu.Unlock()

	// user.IsInsideRoom = false
}

func (r *Room) DeleteRoom() {
	// room := rooms[room_id]
	mu.Lock()
	for name, _ := range r.Users {

		delete(r.Users, name)
	}
	mu.Unlock()
}

func (r *Room) MessageToAll(from User, msg string) {
	// room := rooms[room_id]
	mu.Lock()
	for _, user := range r.Users {
		user.SendMessageInRoom(from, r.Room_name, msg)
	}
	mu.Unlock()
}
func (r *Room) NotificationToAll(msg string) {
	// room := rooms[room_id]
	mu.Lock()
	for _, user := range r.Users {
		user.SendNotification(msg)
	}
	mu.Unlock()
}

func (r *Room) handleChat(message Message) {
	rawMessage := message.Msg
	// sender := message.User

	content := strings.SplitN(rawMessage, " ", 2)
	// cmd := strings.TrimSpace(content[0])

	// switch {
	// case cmd == "/chat":

	msg := strings.TrimSpace(content[1])
	// r.MessageToAll(message.User, msg)
	mu.Lock()
	for _, user := range r.Users {

		// if strings.Compare(user.Name, sender.Name) != 0 {
		user.SendMessageInRoom(message.User, r.Room_name, msg)

		// }
	}
	mu.Unlock()
	// message.User.SendNotification(menu)

	// case cmd == "/list":
	// 	var members string
	// 	mu.Lock()
	// 	for name, _ := range r.Users {
	// 		members += fmt.Sprintf("- %s\n", name)
	// 	}
	// 	mu.Unlock()
	// 	// message.User.SendNotification(menu)
	// case cmd == "/leave":
	// 	message.User.CurrentRoom = nil
	// 	message.User.IsInsideRoom = false

	// 	r.Leave <- message.User

	// default:
	// 	message.User.SendNotification("Unknown Command")
	// 	// message.User.SendNotification(menu)

	// }
}
