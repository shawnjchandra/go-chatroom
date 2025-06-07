package server_module

import (
	"bufio"
	"fmt"
	"net"
)

type User struct {
	Name         string
	Conn         net.Conn
	ListRoom     []string
	Counter      int
	IsInsideRoom bool
	CurrentRoom  *Room
}

func CreateUser(name string, conn net.Conn) User {
	user := User{
		Name:         name,
		Conn:         conn,
		ListRoom:     make([]string, 99),
		Counter:      0,
		IsInsideRoom: false,
		CurrentRoom:  nil,
	}

	return user
}

func (u User) CloseConnection() string {
	err := u.Conn.Close()
	if err != nil {
		return fmt.Sprintf("%s Connection closed\n", u.Name)
	} else {
		return fmt.Sprintf("Failed to close connection for %s\n", u.Name)

	}
}

func (u User) ReceiveMessage() (string, error) {
	reader := bufio.NewReader(u.Conn)
	msg, err := reader.ReadString('\n')

	if err != nil {
		return "", err

	}

	return msg, nil

	// for cName, Conn := range clients {
	// 	if strings.Compare(cName, Name) != 0 {
	// 		Conn.Write([]byte(cName + ":" + msg))

	// 	}
	// }
}

func (u User) SendMessage(from User, msg string) {

	wrapped := fmt.Sprintf("[FROM | %s] : %s\n", from.Name, msg)
	u.Conn.Write([]byte(wrapped))
}

func (u User) SendMessageInRoom(from User, room_name, msg string) {

	wrapped := fmt.Sprintf("[FROM | %s - %s] : %s\n", room_name, from.Name, msg)
	u.Conn.Write([]byte(wrapped))
}

func (u User) SendNotification(msg string) {

	wrapped := fmt.Sprintf("[FROM | SERVER] : %s\n", msg)
	u.Conn.Write([]byte(wrapped))

}
