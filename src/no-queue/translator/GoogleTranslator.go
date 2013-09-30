package Translator

import (
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
	"net"
	"log"
	"net/url"
)

type TranslateJob struct {
	Url, Srctxt, Srclang, Tgttxt, Tgtlang, Echotxt string
	Echo, Kill bool
}

func HandleRequest(conn net.Conn) {
	for {
		var j TranslateJob
		dec := json.NewDecoder(conn)
		err := dec.Decode(&j)

		if err != nil {
			log.Println(err)
			conn.Close()
			return
		}

		if j.Kill == true {
			conn.Close()
			return
		}

		//Encodes the message
		v := url.Values{}
		v.Set("q", j.Srctxt)
		v.Add("client", "t")
		v.Add("text", "")
		v.Add("sl", j.Srclang)
		v.Add("tl", j.Tgtlang)

		//Google's translation address
		s := "http://translate.google.com/translate_a/t?"
		j.Url = s + v.Encode()

		err = Translate(&j, false)

		if err != nil {
			log.Println(err)
		}

		//"Echoes" the translation if desired.
		if j.Echo == true {
			v := url.Values{}
			v.Set("q", j.Tgttxt)
			v.Add("client", "t")
			v.Add("text", "")
			v.Add("sl", j.Tgtlang)
			v.Add("tl", j.Srclang)

			j.Url = s + v.Encode()

			err = Translate(&j, true)

			if err != nil {
				log.Println(err)
			}
		}

		enc := json.NewEncoder(conn)
		err = enc.Encode(j)

		if err != nil {
			log.Println(err)
		}
	}

}

func Translate(request *TranslateJob, echo bool) error {
	//Contact the server.
	resp, err := http.Get(request.Url)

	if err != nil {
		return err
	}

	//Read server's response
	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	var f interface{}
	err = json.Unmarshal(sanitizeReturn(contents, 3), &f)

	if err != nil {
		return err
	}

	//Extract the translated text
	json := f.([]interface{})

	arr := json

	for i := 0; i < 2; i++ {
		s, ok := arr[0].([]interface{})

		if !ok {
			return nil
		}

		arr = s
	}

	if echo == true {
		request.Echotxt = arr[0].(string)
		return nil
	}

	request.Tgttxt = arr[0].(string)
	request.Srclang = json[2].(string)

	return nil
}

func sanitizeReturn(result []byte, iterations int) []byte {
	if(iterations > 1) {
		result = sanitizeReturn(result, iterations - 1)
	}

	str := strings.Replace(string(result), ",,", ",0,", -1)
	return []byte(str)
}
