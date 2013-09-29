package main

import (
	"log"
	"net"
	"./translator"
)

func main() {
	l, err := net.Listen("tcp", "localhost:4242")

	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Fatal(err)
		}

		go Translator.HandleRequest(conn)
	}
}
