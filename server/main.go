package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

var clients = make(map[string]net.Conn)
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

		fmt.Println(conn)
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)

	mu.Lock()
	rawName, _ := reader.ReadString('\n')
	name := strings.TrimSpace(rawName)
	if _, ok := clients[name]; !ok {
		clients[name] = conn
	}
	mu.Unlock()

	// fmt.Printf("Added user : %s", name)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("\n[SERVER - %s] : %s", name, msg)

		for cName, conn := range clients {
			if strings.Compare(cName, name) != 0 {
				conn.Write([]byte(cName + ":" + msg))

			}
		}
	}
}

func closeConnection(conn net.Conn) {
	defer conn.Close()
}
