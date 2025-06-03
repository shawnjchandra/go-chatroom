package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

/*
	Sumber : https://pkg.go.dev/net
*/

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		// handle error
		panic(err)
	}
	// fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	defer conn.Close()

	fmt.Println("Connected to server : ", conn.RemoteAddr())
	// status, err := bufio.NewReader(conn).ReadString('\n')

	go handleServerListen(conn)

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("\nInput name : ")
	scanner.Scan()
	user := scanner.Text() + "\n"

	conn.Write([]byte(user))

	for {

		// fmt.Print("Type Message: ")
		fmt.Print("Type Message: ") // reprint prompt
		scanner.Scan()
		text := scanner.Text() + "\n"

		if strings.Compare(text, "close") == 0 {
			close(conn)
		}

		_, err := conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Error sending message :", err)
			break
		}

	}

	// fmt.Println(status)
	// ...
}

func handleServerListen(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		splitted := strings.Split(msg, ":")
		if err != nil {
			fmt.Println(err)
			continue
		}

		cName := splitted[0]
		splitMsg := splitted[1]

		fmt.Printf("\n[%s] : %s", cName, splitMsg)
		fmt.Print("\nType Message: ") // reprint prompt
	}
}

func close(conn net.Conn) {
	conn.Close()
}
