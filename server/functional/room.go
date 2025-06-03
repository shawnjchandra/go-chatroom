package server_module

import (
	"sync"
)

var mu sync.Mutex

// var rooms = make([]Room, 99)

type Room struct {
	Room_id   int
	Room_name string
	Users     map[string]User
}

func CreateRoom(room_name string, id int) Room {
	// id := len(rooms)
	return Room{
		Room_id:   id,
		Room_name: room_name,
		Users:     make(map[string]User),
	}

}

func (r Room) JoinRoom(user User, room_name string) {
	// room := rooms[room_id]
	r.Users[user.Name] = user

	count := user.Counter
	user.ListRoom[count] = room_name

	user.Counter++
}

func (r Room) DeleteRoom() {
	// room := rooms[room_id]

	for name, _ := range r.Users {

		delete(r.Users, name)
	}
}

func (r Room) MessageToAll(from User, msg string) {
	// room := rooms[room_id]

	for _, user := range r.Users {
		user.SendMessage(from, msg)
	}

}
func (r Room) NotificationToAll(msg string) {
	// room := rooms[room_id]

	for _, user := range r.Users {
		user.SendNotification(msg)
	}

}
