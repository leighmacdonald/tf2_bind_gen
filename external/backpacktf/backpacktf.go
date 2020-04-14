package backpacktf

import (
	"bind_generator/consts"
	"bind_generator/steam"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	urlToken     = "https://backpack.tf/oauth/access_token"
	urlInventory = "https://backpack.tf/api/1.0/inventory/%d/values"
)

var httpClient http.Client

type MarketValue struct {
	// Refined total
	Value float64 `json:"value"`
	// USD
	MarketValue float64 `json:"market_value"`
}

type authToken struct {
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	AccessToken string `json:"access_token"`
}

var token authToken

func Authenticate() error {
	clientId := viper.GetString("backpacktf.client_id")
	if clientId == "" {
		log.Errorf("backpacktf.client_id empty")
		return consts.ErrInvalidConfig
	}
	clientSecret := viper.GetString("backpacktf.client_secret")
	if clientSecret == "" {
		log.Errorf("backpacktf.client_id empty")
		return consts.ErrInvalidConfig
	}
	u := urlToken
	v := url.Values{}
	v.Set("grant_type", "client_credentials")
	v.Set("client_id", clientId)
	v.Set("client_secret", clientSecret)
	v.Set("scope", "read write")
	resp, err := httpClient.PostForm(u, v)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(resp.Body)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("backpack.tf Failed to close resp body: %s", err.Error())
		}
	}()
	var newToken authToken
	if err := json.Unmarshal(b, &newToken); err != nil {
		return err
	}
	log.Debugf("backpack.tf authorization success")
	token = newToken
	return nil
}

func InventoryValues(sid steam.SID64) (MarketValue, error) {
	var mv MarketValue
	u, err := url.Parse(fmt.Sprintf(urlInventory, sid.Int64()))
	if err != nil {
		return MarketValue{}, err
	}
	v := u.Query()
	v.Set("key", sid.String())
	u.RawQuery = v.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return MarketValue{}, err
	}
	req.Header.Set("Authorization", token.AccessToken)
	resp, err := httpClient.Do(req)
	if err != nil {
		return MarketValue{}, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warnf("BPTF Failed to close resp body: %s", err.Error())
		}
	}()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return MarketValue{}, err
	}
	if err := json.Unmarshal(b, &mv); err != nil {
		return MarketValue{}, err
	}
	return mv, nil
}

func init() {
	httpClient = http.Client{Timeout: time.Second}
}
