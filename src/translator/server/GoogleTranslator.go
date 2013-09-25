package googleTranslator
//package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
	"time"
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
}

func Translate(request TranslateJob, retChannel net.Conn) {
	//Contact the server.
	resp, err := http.Get(request.Url)

	if err != nil {
		//Log error
		fmt.Println(err)
		return
	}

	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		//log error
		fmt.Println(err)
		return
	}

	//Parse returned JSON
	var f interface{}
	err = json.Unmarshal(sanitizeReturn(contents, 3), &f)

	if err != nil {
		//log error
		return
	}

	//Extract the translated text
	json := f.([]interface{})

	arr := json

	for i := 0; i < 2; i++ {
		s, ok := arr[0].([]interface{})

		if !ok {
			//log error
			return
		}

		arr = s
	}

	fmt.Println("Text: ", arr[0])
	fmt.Println("Language: ", json[2])

	request.Tgttxt = arr[0].(string)
	request.Srclang = json[2].(string)

	io.Copy(request.Tgttxt, c)
}

func sanitizeReturn(result []byte, iterations int) []byte {
	if(iterations > 1) {
		result = sanitizeReturn(result, iterations - 1)
	}

	str := ToGoString(result)
	str = strings.Replace(str, ",,", ",0,", -1)
	return []byte(str)
}

func main() {
	url := "http://translate.google.com/translate_a/t?q=drag%20if%20you%20want%20to%20work%20on%20this%20all%20night%20ill%20drink%20some%20coffee%20its%20no%20prob&client=t&text=&sl=auto&tl=fr"
	job := TranslateJob{url, "", "", "", ""}
	Translate(job)
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
