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
		Beberapa sumber referensi :
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

		// Terima koneksi
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Println(err)
		}

		fmt.Printf("[ %s ] New Connection has been established\n", logTime())

		// Routine untuk setiap koneksi
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	//handle register
	var user sm.User
	var name string

	// Loop Register
	for {

		rawName, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		name = strings.TrimSpace(rawName)

		/**
		Username tidak bisa kosong atau sudah pernah ada
		Diulang hingga mendapatkan username yang memenuhi
		syarat
		*/
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

	// Notifikasi untuk setiap client / user lainnya
	mu.Lock()
	for _, client := range clients {
		if strings.Compare(client.Name, user.Name) != 0 {
			client.SendNotification(fmt.Sprintf("%s has joined the server", user.Name))
		}
	}
	mu.Unlock()
	conn.Write([]byte(mainMenu))

	// Loop untuk aksi dari user
	for {
		inp, err := reader.ReadString('\n')

		// Condition untuk cek error / force disconnect dsb
		if err != nil {
			res := fmt.Sprintf("%s has forcibly closed the connection\n", user.Name)

			fmt.Printf("[ %s ] %s", logTime(), res)

			// Notifikasi untuk setiap client kalau user A disconnect
			mu.Lock()
			delete(clients, user.Name)
			for _, cl := range clients {
				cl.SendNotification(res)
			}
			mu.Unlock()

			return
		}

		// Trim input dari whitespace
		trimmedInp := strings.TrimSpace(inp)

		// Cek kalau user ada di dalam room atau tidka
		/*
			Kalau ya, pakai menu global
			Kalau tidak ,pakai menu room
		*/
		if !user.IsInsideRoom {

			// Command untuk chat global
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
				// Command untuk cek semua room yang ada

				count := len(rooms)

				if count == 0 {
					conn.Write([]byte("No Rooms Available\n"))
				} else {

					mu.Lock()

					// Write untuk semua room yang ada
					for _, r := range rooms {
						conn.Write([]byte(fmt.Sprintf("%d. %s [%d user]\n", r.Room_id, r.Room_name, len(r.Users))))

					}
					mu.Unlock()

				}
			} else if strings.HasPrefix(trimmedInp, "/create") {
				// Command untuk membuat room

				content := strings.SplitN(trimmedInp, " ", 2)
				// Notifikasi kalau nama roomnya ga dicantumin
				if len(content) <= 1 {
					user.SendNotification("You can't make a <blank> room")
					continue
				}

				room_name := strings.TrimSpace(content[1])

				var ok bool

				mu.Lock()

				count := len(rooms)
				if _, ok = rooms[room_name]; !ok {
					// Buat room kalau di mapping belum ada

					room := sm.CreateRoom(room_name, count)

					// Masukin ke map untuk room yang sudah dibuat dengan key room_name
					rooms[room_name] = room

					// Jalanin room nya sebagai routine terpisah
					go room.Run()

					// Update kondisi dari user
					user.IsInsideRoom = true
					user.CurrentRoom = &room

					// Join user ke room
					user.CurrentRoom.Join <- user
				}

				mu.Unlock()

				/**
				Bisa digabung dengan sebelumnya, tapi dipisahkan
				untuk tujuan mutex, dan karena hanya untuk print ke
				user & server
				*/
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
				// Command untuk join room (yang sudah ada)

				content := strings.SplitN(trimmedInp, " ", 2)

				// Notifikasi kalau belum cantumin nama room
				if len(content) <= 1 {
					user.SendNotification("Hey what's the name...?")
					continue
				}

				room_name := strings.TrimSpace(content[1])

				// Lock dulu
				mu.Lock()

				// Cek kalau di mapping udah ada atau belum
				if r, ok := rooms[room_name]; ok {
					// Langsung unlock karena cuman perlu di pengecekan
					// saja
					mu.Unlock()

					// ganti kondisi user untuk masuk ke room
					user.IsInsideRoom = true
					user.CurrentRoom = &r

					// Join user ke room
					user.CurrentRoom.Join <- user

					res := fmt.Sprintf("%s have joined %s\n", user.Name, room_name)

					conn.Write([]byte(res))
					fmt.Printf("[ %s | %s ] %s\n", logTime(), user.Name, res)

				} else {
					mu.Unlock()
					conn.Write([]byte("The room you're searching for.. doesn't exist\n"))
				}
			} else if strings.HasPrefix(trimmedInp, "/exit") {
				// Command untuk exit dari server

				res := fmt.Sprintf("%s has left the server\n", user.Name)

				// Lock untuk delete user dari mapping client, dan notifikasi ke semua clients
				mu.Lock()
				delete(clients, user.Name)

				for _, cl := range clients {
					cl.SendNotification(res)
				}
				mu.Unlock()

				fmt.Printf("[ %s | %s ] %s", logTime(), user.Name, res)

				return
			} else if strings.HasPrefix(trimmedInp, "/menu") {

				// Command untuk cek menu
				conn.Write([]byte(mainMenu))
			} else {

				// Kalau salah command
				conn.Write([]byte("Unknown command.\n"))
				conn.Write([]byte(mainMenu))

			}

		}

		// Kalau user sudah di dalam room
		if user.IsInsideRoom {

			// Loop untuk command di room
			for {

				inp, err := reader.ReadString('\n')
				if err != nil {
					fmt.Println(err)
				}

				trimmedInp := strings.TrimSpace(inp)

				// Command untuk chat di room
				if strings.HasPrefix(trimmedInp, "/chat") {
					// Bikin message

					newMessage := sm.Message{
						User: user,
						Msg:  inp,
					}

					// Broadcast message ke room
					user.CurrentRoom.Broadcast <- newMessage
				} else if strings.HasPrefix(trimmedInp, "/leave") {
					// Command untuk leave room

					// Untuk logging di server
					res := fmt.Sprintf("%s has left the room %s\n", user.Name, user.CurrentRoom.Room_name)

					// Set kondisi user dan tinggalkan room
					user.IsInsideRoom = false
					user.CurrentRoom.Leave <- user

					fmt.Printf("[ %s | %s ] %s", logTime(), user.Name, res)
					conn.Write([]byte(roomMenu))
					break
				} else if strings.HasPrefix(trimmedInp, "/list") {
					// Ambil semua user di room
					usersInRoom := user.CurrentRoom.Users

					conn.Write([]byte(fmt.Sprint("=== Active User in the room ===\n")))

					// Lock dan ambil semua nama users, lalu di print send ke user
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

// Fungsi untuk menulis log waktu sekarang
func logTime() string {
	t := time.Now()
	formatted := fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d:%04d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), t.Nanosecond())

	return formatted

}
