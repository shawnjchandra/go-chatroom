package main

import (
	"bufio"
	"fmt"
	"net"
	sm "server_module/functional"
	"strings"
	"sync"
	"time"
)

var (
	clients  = make(map[string]sm.User)
	rooms    = make(map[string]sm.Room)
	mu       sync.Mutex
	mainMenu = "Menu: \n1. /all <msg>\n2. /rooms\n3. /create <nama room>\n4. /join <nama room>\n5. /menu\n6. /exit\n"
	roomMenu = "[In-Room Menu]\n" +
		"1. /chat <message>\n" +
		"2. /list\n" +
		"3. /leave\n"
)

func main() {
	/*
		Sumber :
		- https://pkg.go.dev/net
		- https://stackoverflow.com/questions/20234104/how-to-format-current-time-using-a-yyyymmddhhmmss-format
	*/

	ln, err := net.Listen("tcp", ":39563")
	if err != nil {
		// handle error
		fmt.Println("Error on initialization")
	} else {
		fmt.Printf("[ %s ] Server is running\n", logTime())
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Println(err)
		}

		fmt.Printf("[ %s ] New Connection has been established\n", logTime())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	//handle register
	var user sm.User
	var name string

	for {
		rawName, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name = strings.TrimSpace(rawName)

		if name == "" {
			conn.Write([]byte("Name can't be empty, try again\n"))
		} else {

			mu.Lock()
			if _, ok := clients[name]; !ok {
				user = sm.CreateUser(name, conn)
				clients[name] = user
				mu.Unlock()
				break
			} else {
				mu.Unlock()
				conn.Write([]byte("Username is already taken, try again\n"))
			}
		}
		conn.Write([]byte("Input name :\n"))

	}

	fmt.Printf("[ %s ] Added user : %s\n", logTime(), name)

	conn.Write([]byte("Welcome, " + name + "\n"))

	for _, client := range clients {
		if strings.Compare(client.Name, user.Name) != 0 {
			client.SendNotification(fmt.Sprintf("%s has joined the server", user.Name))
		}
	}

	conn.Write([]byte(mainMenu))

	for {
		inp, err := reader.ReadString('\n')
		if err != nil {
			res := fmt.Sprintf("%s has forcibly closed the connection\n", user.Name)

			fmt.Printf("[ %s ] %s", logTime(), res)
			mu.Lock()
			delete(clients, user.Name)
			for _, cl := range clients {
				cl.SendNotification(res)
			}
			mu.Unlock()

			return
		}

		trimmedInp := strings.TrimSpace(inp)

		if !user.IsInsideRoom {

			if strings.HasPrefix(trimmedInp, "/all") {
				content := strings.SplitN(trimmedInp, " ", 2)
				if len(content) <= 1 {
					user.SendNotification("You seem to forgot to write the message...?")
					continue
				}

				msg := strings.TrimSpace(content[1])
				mu.Lock()
				for _, cl := range clients {

					if strings.Compare(cl.Name, user.Name) != 0 {
						cl.SendMessage(user, msg)
					}

				}
				mu.Unlock()
				fmt.Printf("[ %s | %s ] %s\n", logTime(), user.Name, msg)
			} else if strings.HasPrefix(trimmedInp, "/rooms") {
				count := len(rooms)
				if count == 0 {
					// fmt.Println("No Rooms Available")
					conn.Write([]byte("No Rooms Available\n"))
				} else {

					mu.Lock()
					for _, r := range rooms {
						conn.Write([]byte(fmt.Sprintf("%d. %s [%d user]\n", r.Room_id, r.Room_name, len(r.Users))))

					}
					mu.Unlock()

				}
			} else if strings.HasPrefix(trimmedInp, "/create") {
				content := strings.SplitN(trimmedInp, " ", 2)
				if len(content) <= 1 {
					user.SendNotification("You can't make a <blank> room")
					continue
				}

				room_name := strings.TrimSpace(content[1])

				var ok bool

				mu.Lock()

				count := len(rooms)
				if _, ok = rooms[room_name]; !ok {
					room := sm.CreateRoom(room_name, count)
					// room.JoinRoom(user)

					rooms[room_name] = room
					go room.Run()

					user.IsInsideRoom = true
					user.CurrentRoom = &room

					user.CurrentRoom.Join <- user
				}

				mu.Unlock()
				if !ok {
					res := fmt.Sprint("Create and join room " + room_name + "\n")

					conn.Write([]byte(res))
					fmt.Printf("[ %s | %s ] %s", logTime(), user.Name, res)

				} else {
					res := "Failed to create room\n"

					conn.Write([]byte(res))
					fmt.Printf("[ %s | %s ] %s", logTime(), user.Name, res)

				}
			} else if strings.HasPrefix(trimmedInp, "/join") {
				content := strings.SplitN(trimmedInp, " ", 2)
				if len(content) <= 1 {
					user.SendNotification("Hey what's the name...?")
					continue
				}

				room_name := strings.TrimSpace(content[1])

				mu.Lock()
				if r, ok := rooms[room_name]; ok {
					mu.Unlock()

					user.IsInsideRoom = true
					user.CurrentRoom = &r

					user.CurrentRoom.Join <- user

					res := fmt.Sprintf("%s have joined %s\n", user.Name, room_name)

					conn.Write([]byte(res))
					fmt.Printf("[%s | %s ] %s\n", logTime(), user.Name, res)

				} else {
					mu.Unlock()
					conn.Write([]byte("The room you're searching for.. doesn't exist\n"))
				}
			} else if strings.HasPrefix(trimmedInp, "/exit") {
				res := fmt.Sprintf("%s has left the server\n", user.Name)

				mu.Lock()
				delete(clients, user.Name)

				for _, cl := range clients {
					cl.SendNotification(res)
				}
				mu.Unlock()

				fmt.Printf("[ %s | %s ] %s", logTime(), user.Name, res)

				return
			} else if strings.HasPrefix(trimmedInp, "/menu") {
				conn.Write([]byte(mainMenu))
			} else {
				conn.Write([]byte("Unknown command.\n"))
				conn.Write([]byte(mainMenu))

			}

		}

		if user.IsInsideRoom {

			for {

				inp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println(err)
				}

				trimmedInp := strings.TrimSpace(inp)

				if strings.HasPrefix(trimmedInp, "/chat") {
					newMessage := sm.Message{
						User: user,
						Msg:  inp,
					}

					user.CurrentRoom.Broadcast <- newMessage
				} else if strings.HasPrefix(trimmedInp, "/leave") {
					user.IsInsideRoom = false
					user.CurrentRoom.Leave <- user

					conn.Write([]byte(roomMenu))
					break
				} else if strings.HasPrefix(trimmedInp, "/list") {
					usersInRoom := user.CurrentRoom.Users

					conn.Write([]byte(fmt.Sprint("=== Active User in the room ===\n")))
					mu.Lock()
					for _, user := range usersInRoom {
						conn.Write([]byte(fmt.Sprintf("%s\n", user.Name)))
					}
					mu.Unlock()

				} else {
					conn.Write([]byte("Unknown command.\n"))
					conn.Write([]byte(roomMenu))

				}

				fmt.Printf("[ %s | %s ] %s", logTime(), user.Name, inp)
			}

		}

	}
}

func logTime() string {
	t := time.Now()
	formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d:%04d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	return formatted

}
