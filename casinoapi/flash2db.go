package casinoapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gode/log"
	"gode/types"
)

const PathPrefix = "/amfphp/json.php"

const ServiceClient = "Client"
const Service5145 = "casino.slot.line243.BuBuGaoSheng"
const Service5156 = "casino.slot.crash.ZumaEmpire"

type Flash2db struct {
	url string
}

func NewFlash2db(url string) *Flash2db {
	return &Flash2db{url: url}
}

func (f *Flash2db) Call(gt types.GameType, function string, parameters ...interface{}) ([]byte, error) {
	service, err := f.getService(gt, function)
	if err != nil {
		return nil, err
	}
	url := f.url + f.makePath(service, function, parameters...)
	response, err := http.Get(url)
	if err != nil {
		msg := fmt.Sprintf("f2db get error: %v", err)
		log.Print(log.Error, msg)
		return nil, fmt.Errorf(msg)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("flash2db not got code 200")
	}
	defer response.Body.Close()
	//todo: understand what this error means
	content, _ := ioutil.ReadAll(response.Body)

	log.Print(log.Debug, fmt.Sprintf("f2db url: %s", url))
	log.Print(log.Debug, fmt.Sprintf("f2db res: %s", content))

	return content, nil
}

func (f *Flash2db) getService(gameType types.GameType, function string) (string, error) {
	if function == LoginCheck {
		return ServiceClient, nil
	}

	switch gameType {
	case types.GameType(5145):
		return Service5145, nil
	case types.GameType(5156):
		return Service5156, nil
	}

	return "", fmt.Errorf("game type not exsits")
}

func (f *Flash2db) makePath(service, function string, parameters ...interface{}) string {
	b := strings.Builder{}

	b.WriteString(fmt.Sprintf("%s/%s.%s", PathPrefix, service, function))

	for _, p := range parameters {
		b.WriteString(fmt.Sprintf("/%v", p))
	}

	return b.String()
}
