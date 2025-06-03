package main

import (
	"bufio"
	"fmt"
	"net"
	sm "server_module/functional"
	"strings"
	"sync"
)

var clients = make(map[string]sm.User)
var rooms = make(map[string]sm.Room)
var mu sync.Mutex

func main() {
	/*
		Sumber : https://pkg.go.dev/net
	*/

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}

		fmt.Println("New Connection has been established")
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	mu.Lock()
	rawName, _ := reader.ReadString('\n')
	name := strings.TrimSpace(rawName)

	user := sm.CreateUser(name, conn)

	if _, ok := clients[name]; !ok {
		clients[name] = user
	}
	mu.Unlock()

	fmt.Printf("\nAdded user : %s", name)

	conn.Write([]byte("Welcome, " + name + "\n"))
	// conn.Write([]byte("1. Chat to all\n2. Create Room\n3. Join Room\n4. Exit"))

	// Nanti ganti deh wkwkwkwk ,jelek
	conn.Write([]byte("Menu:\n1. /chat_all <msg>\n2. /list_room\n3. /create <nama room>\n4. /join <nama room>\n5. /chat_room <nama room>"))

	for {
		inp, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		trimmedInp := strings.TrimSpace(inp)

		if strings.HasPrefix(trimmedInp, "/chat_all") {
			msg := strings.TrimSpace(strings.SplitN(trimmedInp, " ", 2)[1])
			for _, cl := range clients {

				if strings.Compare(cl.Name, user.Name) != 0 {
					cl.SendMessage(user, msg)
				}
			}
		} else if strings.HasPrefix(trimmedInp, "/list_room") {
			count := len(rooms)
			if count == 0 {
				// fmt.Println("No Rooms Available")
				conn.Write([]byte("No Rooms Available\n"))
			} else {
				for _, r := range rooms {
					conn.Write([]byte(fmt.Sprintf("%d. %s\n", r.Room_id, r.Room_name)))

				}
			}
		} else if strings.HasPrefix(trimmedInp, "/create") {
			room_name := strings.TrimSpace(strings.SplitN(trimmedInp, " ", 2)[1])

			mu.Lock()

			count := len(rooms)
			room := sm.CreateRoom(room_name, count)
			room.JoinRoom(user, room_name)
			rooms[room_name] = room

			mu.Unlock()

			conn.Write([]byte("Create and join room " + room_name))
		} else if strings.HasPrefix(trimmedInp, "/join") {
			room_name := strings.TrimSpace(strings.SplitN(trimmedInp, " ", 2)[1])

			if r, ok := rooms[room_name]; ok {
				conn.Write([]byte("The room you're searching for.. doesn't exist\n"))
			} else {
				mu.Lock()
				r.JoinRoom(user, room_name)
				mu.Unlock()

				conn.Write([]byte(fmt.Sprintf("You have joined %s\n", room_name)))
			}
		} else if strings.HasPrefix(trimmedInp, "/chat_room") {
			splitted := strings.SplitN(trimmedInp, " ", 3)
			room_name := splitted[1]
			msg := splitted[2]

			isFound := false

			mu.Lock()
			for _, r := range rooms {
				if strings.Compare(room_name, r.Room_name) == 0 {
					isFound = true
					r.MessageToAll(user, msg)
				}
			}
			mu.Unlock()
			if !isFound {
				conn.Write([]byte("Room Not Found, Here is your connected room :"))
				for e := range user.ListRoom {
					conn.Write([]byte(fmt.Sprintf("- %s\n", e)))
				}
			}
		} else {
			conn.Write([]byte("Unknown command.\n"))
			conn.Write([]byte("Menu:\n1. /chat_all <msg>\n2. /list_room\n3. /create <nama room>\n4. /join <nama room>\n5. /chat_room <nama room>"))

		}
	}
}
