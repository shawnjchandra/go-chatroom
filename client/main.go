package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Dial server
	conn, err := net.Dial("tcp", "localhost:39563")

	// Kalau ada masalah,hentikan program
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	// Pastikan connection ditutup apabila program berhenti
	defer conn.Close()

	fmt.Println("===== Connected to server =====")

	// Jalankan routine client untuk handle input dari server
	go listenFromServer(conn)

	// Scanner dengan buffer
	scanner := bufio.NewScanner(os.Stdin)

	// Bagian input nama
	fmt.Print("Input name : ")
	scanner.Scan()
	username := scanner.Text()

	// Send nama, dan tutup program kalau ada masalah send
	_, err = conn.Write([]byte(username + "\n"))
	if err != nil {
		fmt.Println("Failed to send name:", err)
		return
	}

	// Loop untuk input client side
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
	// Buat reader sebagai receiver message dari server
	reader := bufio.NewReader(conn)

	// Loop untuk receive dari server
	for {

		// Baca message dari server ,dengan limit pada \n
		msg, err := reader.ReadString('\n')
		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() {
				// Timeout
				fmt.Print(">: ")
				continue
			} else {
				//error atau koneksi disconnect
				fmt.Println("\nDisconnected from server.")
				return
			}
		}

		// Kalau ga ada error / masalah, print message dari server
		msg = strings.TrimSpace(msg)

		fmt.Println(msg)

		fmt.Print("> ")
	}
}
