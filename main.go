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

	go func(out <-chan string) {
		for {
			value := <-out
			log.Fatalln("An external command", value, "stopped the service")
		}
	}(commands)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go func(c net.Conn) {
			_, err = c.Read(command)
			if err != nil {
				return
			}
		}(conn)
		handleConn(conn, commands)
	}
}

func handleConn(c net.Conn, in chan<- string) {
	defer c.Close()
	fmt.Println("A client connected to the server", c.RemoteAddr())
	for {
		_, err := io.WriteString(c, time.Now().Format("15:04:05\n\r"))
		if err != nil {
			fmt.Println("The client", c.RemoteAddr(), "has disconnected")
			return
		}

		fmt.Println("Reading result:", string(command), ".")

		if string(command) == "exit" {
			fmt.Println("Received `exit` command!")
			in <- "exit"
		}

		time.Sleep(1 * time.Second)
	}

}
