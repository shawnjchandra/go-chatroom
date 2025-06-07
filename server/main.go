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
	mainMenu = "Menu: \n1. /all <msg>\n2. /rooms\n3. /create <nama room>\n4. /join <nama room>\n5. /menu\n"
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
		}

		fmt.Printf("[ %s ] New Connection has been established\n", logTime())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	rawName, _ := reader.ReadString('\n')

	name := strings.TrimSpace(rawName)

	user := sm.CreateUser(name, conn)

	mu.Lock()
	if _, ok := clients[name]; !ok {
		clients[name] = user
	}

	mu.Unlock()

	fmt.Printf("[ %s ] Added user : %s\n", logTime(), name)

	conn.Write([]byte("Welcome, " + name + "\n"))
	// conn.Write([]byte("1. Chat to all\n2. Create Room\n3. Join Room\n4. Exit"))

	// Nanti ganti deh wkwkwkwk ,jelek
	conn.Write([]byte(mainMenu))

	for {
		inp, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		trimmedInp := strings.TrimSpace(inp)

		if !user.IsInsideRoom {

			if strings.HasPrefix(trimmedInp, "/all") {
				msg := strings.TrimSpace(strings.SplitN(trimmedInp, " ", 2)[1])
				mu.Lock()
				for _, cl := range clients {

					if strings.Compare(cl.Name, user.Name) != 0 {
						cl.SendMessage(user, msg)
						fmt.Printf("[ %s | %s ] %s\n", logTime(), user.Name, msg)
					}

				}
				mu.Unlock()
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
				room_name := strings.TrimSpace(strings.SplitN(trimmedInp, " ", 2)[1])

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
					fmt.Printf("[ %s | %s ] %sx", logTime(), user.Name, res)

				} else {
					res := "Failed to create room\n"

					conn.Write([]byte(res))
					fmt.Printf("[ %s | %s ] %s", logTime(), user.Name, res)

				}
			} else if strings.HasPrefix(trimmedInp, "/join") {
				room_name := strings.TrimSpace(strings.SplitN(trimmedInp, " ", 2)[1])
				// fmt.Println("test", room_name)

				if r, ok := rooms[room_name]; ok {

					// mu.Lock()
					// r.JoinRoom(user)
					// mu.Unlock()

					user.IsInsideRoom = true
					user.CurrentRoom = &r

					user.CurrentRoom.Join <- user

					res := fmt.Sprintf("You have joined %s\n", room_name)

					conn.Write([]byte(res))
					fmt.Printf("[%s | %s ] %s", logTime(), user.Name, res)

				} else if strings.HasPrefix(trimmedInp, "/menu") {
					conn.Write([]byte(mainMenu))
				} else {
					conn.Write([]byte("The room you're searching for.. doesn't exist\n"))
				}
			} else {
				conn.Write([]byte("Unknown command.\n"))
				conn.Write([]byte(mainMenu))

			}

			// conn.Write([]byte("Menu: \n1. /all <msg>\n2. /list_room\n3. /create <nama room>\n4. /join <nama room>\n"))

		}

		if user.IsInsideRoom {
			fmt.Println("in a room")
			// conn.Write([]byte(roomMenu))
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
	formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	return formatted

}
