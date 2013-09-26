package Translator

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
	"time"
	"io"
	"net"
	"log"
)

type TranslateError struct {
	When time.Time
	Message string
}

func (e *TranslateError) Error() string {
	return fmt.Sprintf("Error: %v, %s", e.When, e.Message)
}

type TranslateJob struct {
	Url, Srctxt, Srclang, Tgttxt, Tgtlang string
	Kill bool
}

func HandleRequest(conn net.Conn) {
	for {
		var j TranslateJob
		dec := json.NewDecoder(conn)
		err := dec.Decode(&j)

		if err != nil {
			log.Println("Problem converted the JSON.", err)
			conn.Close()
			return
		}

		if j.Kill == true {
			conn.Close()
			return
		}

		Translate(&j)

		if err != nil {
			log.Println(err)
		}

		io.WriteString(conn, j.Tgttxt)
	}

}

func Translate(request *TranslateJob) error {
	//Contact the server.
	resp, err := http.Get(request.Url)

	if err != nil {
		log.Println("Unable to call translation service.", err)
		return err
	}

	//Read server's response
	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("Unable to read the server's response.", err)
		return err
	}

	var f interface{}
	err = json.Unmarshal(sanitizeReturn(contents, 3), &f)

	if err != nil {
		log.Println("Unable to parse the translation.", err)
		return err
	}

	//Extract the translated text
	json := f.([]interface{})

	arr := json

	for i := 0; i < 2; i++ {
		s, ok := arr[0].([]interface{})

		if !ok {
			log.Println("Error while reading the JSON.")
			return
		}

		arr = s
	}

	request.Tgttxt = arr[0].(string)
	request.Srclang = json[2].(string)
}

func sanitizeReturn(result []byte, iterations int) []byte {
	if(iterations > 1) {
		result = sanitizeReturn(result, iterations - 1)
	}

	str := ToGoString(result)
	str = strings.Replace(str, ",,", ",0,", -1)
	return []byte(str)
}

func ToGoString(c []byte) string {
	n := -1

	for i, b := range c {
		if b == 0 {
			break
		}
		n = i
	}

	return string(c[:n+1])
}
