package casinoapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const urlPrefix = "/amfphp/json.php"

type Flash2db struct {
	host string
}

func NewFlash2db(host string) *Flash2db {
	return &Flash2db{host: host}
}

func (f *Flash2db) Call(service, function string, parameters ...interface{}) ([]byte, error) {
	response, _ := http.Get(f.host + f.makeURL(service, function, parameters...))
	defer response.Body.Close()
	content, _ := ioutil.ReadAll(response.Body)

	return content, nil
}

func (f *Flash2db) makeURL(service, function string, parameters ...interface{}) (URL string) {
	b := strings.Builder{}

	b.WriteString(fmt.Sprintf("%s/%s.%s", urlPrefix, service, function))

	for _, p := range parameters {
		b.WriteString(fmt.Sprintf("/%v", p))
	}

	return b.String()
}
