package main

import (
//	"io"
	"log"
	"net"
	"./googleTranslator"
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

	url := "http://translate.google.com/translate_a/t?q=drag%20if%20you%20want%20to%20work%20on%20this%20all%20night%20ill%20drink%20some%20coffee%20its%20no%20prob&client=t&text=&sl=auto&tl=fr"
	job := googleTranslator.TranslateJob{url, "", "", "", ""}

		go googleTranslator.Translate(job, conn)
//		go func(c net.Conn) {
//			io.Copy(c, c)
//			c.Close()
//		}(conn)
	}
}
