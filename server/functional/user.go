package server_module

import (
	"fmt"
	"net"
)

type User struct {
	Name         string
	Conn         net.Conn
	IsInsideRoom bool
	CurrentRoom  *Room
}

func CreateUser(name string, conn net.Conn) User {
	user := User{
		Name:         name,
		Conn:         conn,
		IsInsideRoom: false,
		CurrentRoom:  nil,
	}

	return user
}

func (u User) CloseConnection() {
	u.Conn.Close()

}

// Message global
func (u User) SendMessage(from User, msg string) {

	wrapped := fmt.Sprintf("[ FROM | %s ] : %s\n", from.Name, msg)
	u.Conn.Write([]byte(wrapped))
}

// Message di dalam room
func (u User) SendMessageInRoom(from User, room_name, msg string) {

	wrapped := fmt.Sprintf("[ FROM | %s - %s ] : %s\n", room_name, from.Name, msg)
	u.Conn.Write([]byte(wrapped))
}

// Notifikasi dari server
func (u User) SendNotification(msg string) {

	wrapped := fmt.Sprintf("[ FROM | SERVER ] : %s\n", msg)
	u.Conn.Write([]byte(wrapped))

}
