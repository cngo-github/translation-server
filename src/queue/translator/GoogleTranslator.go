package translator

import (
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
	"log"
	"net/url"
)

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

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:24.0) Gecko/20100101 Firefox/24.0")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Println("Failed to read HTTP body.")
		return err
	}

	var f interface{}
	retJson := sanitizeReturn(string(contents), 3)
	err = json.Unmarshal([]byte(retJson), &f)

	if err != nil {
		log.Println("JSON failed to unmarshel.")
		return err
	}

	//Extract the translated text
	json := f.([]interface{})

	arr := json

	for i := 0; i < 1; i++ {
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

	var txt string

	for i := 0; i < cap(arr); i++ {
		arrText := arr[i].([]interface{})
		txt = txt + arrText[0].(string) + " "
	}

	request.Tgttxt = txt
	request.Srclang = json[2].(string)

	return nil
}

func sanitizeReturn(result string, iterations int) string {
	if(iterations > 1) {
		result = sanitizeReturn(result, iterations - 1)
	}

	for iterations >= 0 {
		result = strings.Replace(result, ",,", ",0,", -1)
		result = strings.Replace(result, "[,", "[0,", -1)
		iterations = iterations - 1
	}

	return result
}
