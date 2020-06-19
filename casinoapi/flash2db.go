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
	response, err := http.Get(f.host + f.makeURL(service, function, parameters...))
	if err != nil {
		return nil, fmt.Errorf("flash2db get error %v", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("flash2db not got code 200")
	}
	defer response.Body.Close()
	//todo: understand what this error means
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
