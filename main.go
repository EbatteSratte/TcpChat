package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

type Client struct {
	Conn net.Conn
	Addr string
}

var (
	clients   = make(map[net.Conn]*Client)
	mu        sync.Mutex
	broadcast = make(chan []byte)
)

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
	defer listener.Close()

	fmt.Println("Сервер запущен на :8080")

	go handleMessages()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Ошибка подключения клиента:", err)
			continue
		}

		mu.Lock()
		clients[conn] = &Client{Conn: conn, Addr: conn.RemoteAddr().String()}
		mu.Unlock()

		fmt.Printf("Клиент подключился: %s\n", conn.RemoteAddr().String())
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer func() {
		mu.Lock()
		delete(clients, conn)
		mu.Unlock()
		conn.Close()
		fmt.Printf("Клиент отключился: %s\n", conn.RemoteAddr().String())
	}()

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		fmt.Printf("Получено от %s: %s", conn.RemoteAddr(), message)
		broadcast <- []byte(message)
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		mu.Lock()
		for _, client := range clients {
			_, err := client.Conn.Write(msg)
			if err != nil {
				client.Conn.Close()
				delete(clients, client.Conn)
			}
		}
		mu.Unlock()
	}
}
