package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var wg sync.WaitGroup

func main() {
	conn, err := net.Dial("tcp", "localhost:39563")
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("===== Connected to server =====")

	// Handle incoming messages from server
	go listenFromServer(conn)

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Input name: ")
	scanner.Scan()
	username := scanner.Text()
	_, err = conn.Write([]byte(username + "\n"))
	if err != nil {
		fmt.Println("Failed to send name:", err)
		return
	}

	for {

		fmt.Print("> ")
		scanner.Scan()
		text := scanner.Text()

		if strings.TrimSpace(text) == "close" {
			fmt.Println("Closing connection...")
			return
		}

		_, err := conn.Write([]byte(text + "\n"))
		if err != nil {
			fmt.Println("Error sending message:", err)
			break
		}
	}
}

func listenFromServer(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		// Set a read deadline 1 second from now
		// conn.SetReadDeadline(time.Now().Add(5 * time.Second))

		msg, err := reader.ReadString('\n')
		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() {
				// Timeout reached, no data
				fmt.Print(">: ")
				continue
			} else {
				// Real error or connection closed
				fmt.Println("\nDisconnected from server.")
				return
			}
		}

		msg = strings.TrimSpace(msg)

		fmt.Println(msg)
		// parts := strings.SplitN(msg, ":", 2)

		// if len(parts) == 2 {
		// 	sender := strings.TrimSpace(parts[0])
		// 	content := strings.TrimSpace(parts[1])
		// 	fmt.Printf("%s: %s\n", sender, content)
		// } else {
		// 	fmt.Println(msg)
		// }

		fmt.Print("> ")
	}
}
