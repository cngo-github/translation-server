//package GoogleTranslator
package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
	"encoding/json"
)

type TranslationResult struct {
        url, srctxt, srclang, tgttxt, tgtlang string
}

func Translate() {
	resp, err := 
http.Get("http://translate.google.com/translate_a/t?q=drag%20if%20you%20want%20to%20work%20on%20this%20all%20night%20ill%20drink%20some%20coffee%20its%20no%20prob&client=t&text=&sl=auto&tl=fr")

	if err != nil {
		fmt.Println(err)
		return
	}

	contents, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		return
	}

	parseReturn(contents)
}

func parseReturn(result []byte) {
	result = sanitizeReturn(result, 3)

	var f interface{}
	err2 := json.Unmarshal(result, &f)

                if err2 == nil {  
                        m := f.([]interface{})

                        for k, v := range m {
                                switch vv := v.(type) {
                                        case string:
                                                fmt.Println(k, " is string ", vv)
                                        case int:
                                                fmt.Println(k, " is int ", vv)
                                        case []interface{}:   
                                                fmt.Println(k, " is an array ")
                                        default: 
                                                fmt.Println(k, " is unknown.")
                                }
                        }
                        fmt.Printf("Unmarshelled: %s\n", f)
                }
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
	Translate()
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
