package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

var command []byte

func main() {
	command = make([]byte, 4)
	commands := make(chan string, 1)

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

	// Запускаем читалку канала
	go func(out <-chan string) {
		for {
			value := <-out
			if value == "exit" {
				log.Fatalln("An external command", value, "stopped the service")
			}
		}
	}(commands)

	// Раз есть соединение, запускаем читалку сокета
	go func(c net.Conn, in chan<- string) {
		for {
			fmt.Println("Reading the socket...")
			_, err = c.Read(command)
			if err != nil {
				log.Println(err)
				return
			}

			if string(command) == "exit" {
				fmt.Println("Received `exit` command!")
				in <- "exit"
			}
		}
	}(conn, commands)

	// Запускаем отправлялку времени
	handleConn(conn)
}

func handleConn(c net.Conn) {
	defer c.Close()
	fmt.Println("A client connected to the server", c.RemoteAddr())
	for {
		_, err := io.WriteString(c, time.Now().Format("15:04:05\n\r"))
		if err != nil {
			fmt.Println("The client", c.RemoteAddr(), "has disconnected")
			return
		}

		time.Sleep(5 * time.Second)
	}
}
