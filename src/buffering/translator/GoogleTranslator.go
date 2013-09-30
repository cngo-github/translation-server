package translator

import (
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
	"log"
	"net/url"
//	"bytes"
//	"code.google.com/p/go-charset/charset"
//	"fmt"
//	"sjisconv"
//	"unicode/utf16"
//	"unicode/utf8"
//	"io"
//	"os"

	"code.google.com/p/go.text/encoding/charmap"
	"code.google.com/p/go.text/encoding/japanese"
	"code.google.com/p/go.text/encoding/korean"
	"code.google.com/p/go.text/encoding/simplifiedchinese"
	"code.google.com/p/go.text/encoding/traditionalchinese"
//	"code.google.com/p/go.text/encoding/unicode"
	"code.google.com/p/go.text/transform"
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

	var tr (*transform.Reader)

	switch resp.Header.Get("Content-Language") {
		case "ar":
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ", GBK -> UTF-8.")
			tr = transform.NewReader(resp.Body, charmap.ISO8859_6.NewDecoder())
		case "ja":
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ", ShiftJIS -> UTF-8")
			tr = transform.NewReader(resp.Body, japanese.ShiftJIS.NewDecoder())
		case "en":
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ", no conversions.")
		case "ko":
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ", EUCKR -> UTF-8.")
			tr = transform.NewReader(resp.Body, korean.EUCKR.NewDecoder())
		case "ru", "bg", "uk":
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ", EUCKR -> KOI8R.")
			tr = transform.NewReader(resp.Body, charmap.KOI8R.NewDecoder())
		case "zh-CN":
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ", GBK -> UTF-8.")
			tr = transform.NewReader(resp.Body, simplifiedchinese.GBK.NewDecoder())
		case "zh-TW", "th":
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ", Big5 -> UTF-8.")
			tr = transform.NewReader(resp.Body, traditionalchinese.Big5.NewDecoder())
		default:
			log.Println("Langauge: " + resp.Header.Get("Content-Language") + ".  Using ISO-8859-15 -> UTF-8")
			tr = transform.NewReader(resp.Body, charmap.Windows1252.NewDecoder())
	}

	if tr == nil {
		log.Println("Failed to convert the JSON")
		return nil
	}

	contents, err := ioutil.ReadAll(tr)

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
