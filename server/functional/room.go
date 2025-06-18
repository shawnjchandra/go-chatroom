package server_module

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

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

	// Channel untuk langsung dapat diterima antar routine
	Join      chan User
	Leave     chan User
	Broadcast chan Message
}

func CreateRoom(room_name string, id int) Room {

	return Room{
		Room_id:   id,
		Room_name: room_name,
		Users:     make(map[string]User),

		// Set ke 99 untuk mensimulasikan apabila terdapat banyak user
		Join:      make(chan User, 99),
		Leave:     make(chan User, 99),
		Broadcast: make(chan Message, 99),
	}
}

// Method utama untuk menjalankan routine room
func (r *Room) Run() {

	// Loop selama room masih ada
	for {

		select {
		// Kasus untuk join, dilempar ke method join, dan notifikasi ke user untuk menu pertama kali
		case u := <-r.Join:
			r.JoinRoom(u)
			u.SendNotification(roomMenu)

		// Kasus untuk leave room, langsung dilempar ke method leave
		case u := <-r.Leave:
			r.LeaveRoom(u)

		// Kasus untuk broadcast, diberikan ke method handleChat
		case message := <-r.Broadcast:

			r.handleChat(message)

		}
	}
}

func (r *Room) JoinRoom(user User) error {
	// Lock dulu, dan langsung unlock setelah selesai masuk ke map
	mu.Lock()
	if _, ok := r.Users[user.Name]; !ok {
		// Masukin ke mapping
		r.Users[user.Name] = user
		mu.Unlock()

		// Notifikasi ke semua kalau ada user yang join
		r.NotificationToAll(fmt.Sprintf("%s has joined the room %s", user.Name, r.Room_name))

		return nil
	}
	mu.Unlock()
	return errors.New("can't join room")

}

func (r *Room) LeaveRoom(user User) {
	notif := fmt.Sprintf("%s has left the room %s", user.Name, r.Room_name)
	r.NotificationToAll(notif)

	// Lock, delete user dari mapping users internal , lalu unlock mutex
	mu.Lock()
	delete(r.Users, user.Name)
	mu.Unlock()

}

// Method untuk message (dari user ke user)
func (r *Room) MessageToAll(from User, msg string) {

	mu.Lock()
	for _, user := range r.Users {
		user.SendMessageInRoom(from, r.Room_name, msg)
	}
	mu.Unlock()
}

// Method untuk notifikasi (dari server ke user)
func (r *Room) NotificationToAll(msg string) {

	mu.Lock()
	for _, user := range r.Users {
		user.SendNotification(msg)
	}
	mu.Unlock()
}

// Method untuk handle chat
func (r *Room) handleChat(message Message) {
	rawMessage := message.Msg

	// Split contentnya
	content := strings.SplitN(rawMessage, " ", 2)

	// Hilangin whitespace yang ada di message
	msg := strings.TrimSpace(content[1])

	// Lock, send message, unlock
	r.MessageToAll(message.User, msg)

}
