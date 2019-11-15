package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	go io.Copy(os.Stdout, conn) // игнорируем ошибки

	console := []byte{}
	for {
		console = []byte{}
		fmt.Scan(&console)
		if "exit" == string(console) {
			_, err := io.WriteString(conn,"exit")
			if err != nil {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
}