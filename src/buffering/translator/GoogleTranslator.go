package translator

import (
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
	"log"
	"net/url"
	"bytes"
	"code.google.com/p/go-charset/charset"
	"fmt"
	"sjisconv"
	"unicode/utf16"
	"unicode/utf8"
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

//	log.Println(request.Tgtlang)
	log.Println([]byte(request.Tgttxt))
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

	fmt.Println(contents)

	switch resp.Header.Get("Content-Language") {
		case "ja":
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ".  Using ShiftJIS -> UTF-8")
			contents = shiftJISToUTF8(contents)
		case "en":
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ".  Doing no conversions.")
			contents = iso88591ToUTF8(contents)
//		case "zh-TW", "zh-CH":
//			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ".  Using BIG-5 -> UTF-8.")
//			contents, err = big5ToUTF8(contents)
		default:
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ".  Using ISO-8859-15 -> UTF-8")
			contents, err = iso885915ToUTF8(contents)
	}

//	fmt.Printf("%s", resp)
//	contents, err = iso885915ToUTF8(contents)

//	contents = shiftJISToUTF8(contents)

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

func iso885915ToUTF8(arr []byte) ([]byte, error) {
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

func iso88591ToUTF8(input []byte) []byte {
	// - ISO-8859-1 bytes match Unicode code points
	// - All runes <128 correspond to ASCII, same as in UTF-8
	// - All runes >128 in ISO-8859-1 encode as 2 bytes in UTF-8
	res := make([]byte, len(input)*2)

	var j int
	for _, b := range input {
		if b <= 128 {
			res[j] = b
			j += 1
		} else {
			if b >= 192 {
				res[j] = 195
				res[j+1] = b - 64
			} else {
				res[j] = 194
				res[j+1] = b
			}

			j += 2
		}
	}

	return res[:j]
}

func big5ToUTF8(arr []byte) ([]byte, error) {
	r, err := charset.NewReader("BIG5", bytes.NewReader(arr))

	if err != nil {
		return nil, err
	}

	result, err := ioutil.ReadAll(r)

	if err != nil {
		return nil, err
	}

	return result, nil
}


func shiftJISToUTF8(input []byte) []byte {
	var strBuf string = ""

	Utf16str := sjisconv.SjistoUtf16(input)
	runes  := utf16.Decode(Utf16str)
	buf  := make([]byte, 6)

	for i := 0; i < len(runes); i++ {
		N := utf8.EncodeRune(buf, runes[i])
		strBuf += string(buf[0:N])
	}

	return []byte(strBuf)
}
