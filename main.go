package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

var message []byte

var text string
var name string

var connections []*net.Conn

var users map[string]string

func main() {
	message = make([]byte, 80)

	messages := make(chan string, 1)

	users = make(map[string]string, 10)

	fmt.Println("The server has started")

	listener, err := net.Listen("tcp", "localhost:9000")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			fmt.Println("Ждём соединения")
			conn, err := listener.Accept()
			if err != nil {
				log.Print(err)
			}
			fmt.Println("A new user has connected to the server:", conn.RemoteAddr())
			connections = append(connections, &conn) // Переделать на канал!
			fmt.Println(connections)

			// Раз есть соединение, запускаем читалку сокета
			go func(c net.Conn, in chan<- string) {
				for {
					fmt.Println("Reading the socket...")
					_, err = c.Read(message)
					if err != nil {
						log.Println(err)
						c.Close()
					}

					in <- string(message)
					message = make([]byte, 80) // Обнуляем
				}
			}(conn, messages)
		}
	}()

	// Запускаем читалку канала
	for {
		value := <-messages
		name, text = DecodeByteSlice([]byte(value))
		log.Println("A message from user", name, "received:", text)

		// Broadcaster
		for _, conn := range connections {
			_, err := io.WriteString(*conn, value)
			if err != nil {
				fmt.Println("The client", (*conn).RemoteAddr(), "has disconnected")
				return
			}
		}
	}
}

func DecodeByteSlice(byteSlice []byte) (text1, text2 string) {
	nameLen1 := int(byteSlice[0])
	text1 = string(byteSlice[1 : nameLen1+1])
	nameLen2 := int(byteSlice[nameLen1+1])
	text2 = string(byteSlice[nameLen1+2 : nameLen1+2+nameLen2])
	return
}
