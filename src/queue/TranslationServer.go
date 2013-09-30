package main

import (
	"log"
	"net"
	"./translator"
	"encoding/json"
)

func main() {
	l, err := net.Listen("tcp", "localhost:4242")

	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		queue := make(chan translator.TranslateJob, 100)

		if err != nil {
			log.Fatal(err)
		}

		go processRequest(conn, queue)
	}
}

func processRequest(conn net.Conn, queue chan translator.TranslateJob) {
	for {
		var j translator.TranslateJob
		dec := json.NewDecoder(conn)
		err := dec.Decode(&j)

		if err == nil {
			if j.Kill == true {
				log.Println("Killing connection.")
				conn.Close()
				return
			} else if j.Read == true {
				log.Println("Reading returns.")
				enc := json.NewEncoder(conn)
				err := enc.Encode(<- queue)

				if err != nil {
					log.Println(err)
				}
			} else {
				log.Println("Writing Job.")
				go translator.HandleRequest(j, queue)
			}
		} else {
			log.Println("Unable to read incoming request")
		}
	}
}
