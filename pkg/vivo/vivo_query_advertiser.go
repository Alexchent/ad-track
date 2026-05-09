package vivo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type AdvertiserQueryRequest struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
}

type AdvertiserQueryResponse struct {
	Code    int64          `json:"code,omitempty"`
	Message string         `json:"message,omitempty"`
	Data    AdvertiserInfo `json:"data,omitempty"`
}

type AdvertiserInfo struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	CompanyName string `json:"companyName"`
	IdCardNo    string `json:"idCardNo"`
	ShowName    string `json:"showName"`
}

func (a *AdService) QueryAdvertiser(token string) (*AdvertiserQueryResponse, error) {
	return queryAdvertiser(a.c.Host, token)
}

func queryAdvertiser(host, accessToken string) (*AdvertiserQueryResponse, error) {
	ms := time.Now().UnixNano() / 1e6
	nonce := MakeNonce()

	url := fmt.Sprintf(buildMarketURL(host, vivoAdvertiserQueryURLFormat), accessToken, ms, nonce)
	//fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	result, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	//fmt.Println(string(result))

	response := &AdvertiserQueryResponse{}
	if err := json.Unmarshal(result, response); err != nil {
		return nil, err
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("queryAdvertiser error: %s", response.Message)
	}

	return response, nil
}
