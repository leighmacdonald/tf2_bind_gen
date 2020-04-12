package cryptocompare

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var httpClient http.Client

type Prices struct {
	BTC struct {
		USD float64 `json:"USD"`
		EUR float64 `json:"EUR"`
	} `json:"BTC"`
	ETH struct {
		USD float64 `json:"USD"`
		EUR float64 `json:"EUR"`
	} `json:"ETH"`
}

func GetSymbols(key string, s []string, c []string) (Prices, error) {
	var p Prices
	// https://min-api.cryptocompare.com/data/pricemulti?fsyms=BTC,ETH&tsyms=USD
	u, err := url.Parse("https://min-api.cryptocompare.com/data/pricemulti")
	if err != nil {
		log.Warnf("Failed to setup request url: %s", err.Error())
		return p, err
	}
	q := u.Query()
	for _, ss := range s {
		q.Add("fsyms", ss)
	}
	for _, ss := range c {
		q.Add("tsyms", ss)
	}
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Warnf("Failed to setup request: %s", err.Error())
		return p, err
	}
	req.Header.Set("authorization", fmt.Sprintf("Apikey %s", key))
	resp, er := httpClient.Do(req)
	if er != nil {
		log.Warnf("Failed to make cryptocompare api request: %s", er.Error())
		return p, err
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Failed to read response body: %s", err.Error())
		return p, err
	}
	if err := json.Unmarshal(b, &p); err != nil {
		return p, err
	}
	return p, nil
}

func init() {
	httpClient = http.Client{Timeout: time.Duration(1 * time.Second)}
}
