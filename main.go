package main

import (
	"fmt"
	"io"
	"log"
	"net"
)

var connections []*net.Conn // Временный колхоз

func main() {
	messages := make(chan string, 1)

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

				var message []byte // Сюда мы будем временно записывать массив байтов

				for {
					fmt.Println("Reading the socket...")
					message = make([]byte, 80) // Обнуляем
					_, err = c.Read(message)
					if err != nil {
						log.Println(err)
						c.Close()
					}

					in <- string(message)
				}
			}(conn, messages)
		}
	}()

	// Запускаем читалку канала

	var text string // Переменная хранит текст сообщения
	var name string // Переменная хранит имя пользователя

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

// Функция декодирует сообщения для вывода
func DecodeByteSlice(byteSlice []byte) (text1, text2 string) {
	nameLen1 := int(byteSlice[0])
	text1 = string(byteSlice[1 : nameLen1+1])
	nameLen2 := int(byteSlice[nameLen1+1])
	text2 = string(byteSlice[nameLen1+2 : nameLen1+2+nameLen2])
	return
}