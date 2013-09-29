package translator

import (
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
	"log"
	"net/url"
//	"fmt"
	"bytes"
	"code.google.com/p/go-charset/charset"
)
import _  "code.google.com/p/go-charset/data"

type TranslateJob struct {
	Srctxt, Srclang, Tgttxt, Tgtlang, Echotxt, Channel, User string
	Echo, Kill, Read, Outgoing bool
}

func HandleRequest(request TranslateJob, queue chan TranslateJob) {
	//Encodes the message
	v := url.Values{}
	v.Set("q", request.Srctxt)
	v.Add("client", "t")
	v.Add("text", "")
	v.Add("sl", request.Srclang)
	v.Add("tl", request.Tgtlang)

	//Google's translation address
	s := "http://translate.google.com/translate_a/t?"
	serverUrl := s + v.Encode()

	err := RunTranslation(serverUrl, false, &request)

	if err != nil {
		//Unable to run translation
		log.Println(err)
		return
	}

	log.Println(request.Tgttxt)

	if request.Srctxt == request.Tgttxt {
		//Translation failed or it was the same.
		log.Println("Translation failed or never occurred.")
		queue <- request
		return
	}

	//"Echoes" the translation if desired.
	if request.Echo == true {
		v := url.Values{}
		v.Set("q", request.Tgttxt)
		v.Add("client", "t")
		v.Add("text", "")
		v.Add("sl", request.Tgtlang)
		v.Add("tl", request.Srclang)

		serverUrl := s + v.Encode()

		err := RunTranslation(serverUrl, true, &request)

		if err != nil {
			//Echo failed.
			log.Println(err)
			request.Echo = false
		}
	}

	log.Println("Returning Job.")
	queue <- request
}

func RunTranslation(url string, echo bool, request *TranslateJob) error {
	//Contact the server.
	log.Println("Opening URL: " + url)
	resp, err := http.Get(url)

	if err != nil {
		return err
	}

	//Read server's response
	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	contents, err = iso885915toUtf8(contents)

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

func iso885915toUtf8(arr []byte) ([]byte, error) {
	r, err := charset.NewReader("ISO-8859-15", bytes.NewReader(arr))

	if err != nil {
		return nil, err
	}

	result, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, err
	}

	return result, nil
}
