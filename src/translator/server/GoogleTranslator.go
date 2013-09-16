//package GoogleTranslator
package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"strings"
)

type TranslationResult struct {
        url, srctxt, srclang, tgttxt, tgtlang string
}

func Translate() {
	resp, err := 
http.Get("http://translate.google.com/translate_a/t?q=drag%20if%20you%20want%20to%20work%20on%20this%20all%20night%20ill%20drink%20some%20coffee%20its%20no%20prob&client=t&text=&sl=auto&tl=fr")
	if err == nil {
		contents, err := ioutil.ReadAll(resp.Body)

		if err == nil {
			fmt.Printf("Response: %s\n", contents)
			fmt.Println(contents[0])
//			parseTranslation(contents)
		}
	}
}

func parseTranslation(result string) {
	spl := strings.Split(result, "]")
	fmt.Println(spl)
}

func main() {
	Translate()
}
