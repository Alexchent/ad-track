package vivo

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type VivoAdvertiserQueryRequest struct {
	PageIndex int `json:"pageIndex"`
	PageSize  int `json:"pageSize"`
}

type VivoAdvertiserQueryResponse struct {
	Code    int64              `json:"code,omitempty"`
	Message string             `json:"message,omitempty"`
	Data    VivoAdvertiserInfo `json:"data,omitempty"`
}

type VivoAdvertiserInfo struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	CompanyName string `json:"companyName"`
	IdCardNo    string `json:"idCardNo"`
	ShowName    string `json:"showName"`
}

func queryVivoAdvertiser(accessToken string) (*VivoAdvertiserQueryResponse, error) {
	ms := time.Now().UnixNano() / 1e6
	qid := QidWithUnixTime()
	nonce := fmt.Sprintf("%x", md5.Sum([]byte(qid)))

	url := fmt.Sprintf(VivoAdvertiserQueryUrl, accessToken, ms, nonce)
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

	response := &VivoAdvertiserQueryResponse{}
	if err := json.Unmarshal(result, response); err != nil {
		return nil, err
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("queryVivoAdvertiser error: %s", response.Message)
	}

	return response, nil
}
