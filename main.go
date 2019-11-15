package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

var message []byte

func main() {
	message = make([]byte, 40)
	messages := make(chan string, 1)

	fmt.Println("The server has started")

	listener, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		log.Fatal(err)
	}

	// Ждём соединения (пока пусть 1 клиент)
	conn, err := listener.Accept()
	if err != nil {
		log.Print(err)
	}

	// Раз есть соединение, запускаем читалку сокета
	go func(c net.Conn, in chan<- string) {
		for {
			fmt.Println("Reading the socket...")
			_, err = c.Read(message)
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Println("Считали из сокета:", string(message))
			in <- strings.TrimSpace(string(message))
			message = make([]byte, 40) // Обнуляем
		}
	}(conn, messages)

	fmt.Println("A client connected to the server", conn.RemoteAddr())

	// Запускаем читалку канала
	for {
		value := <-messages
		log.Println("A message received:", value)
		_, err := io.WriteString(conn, value)
		if err != nil {
			fmt.Println("The client", conn.RemoteAddr(), "has disconnected")
			return
		}
	}
}
