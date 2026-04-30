package vivo

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type SummaryQueryRequest struct {
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	PageIndex   int    `json:"pageIndex"`
	PageSize    int    `json:"pageSize"`
	SummaryType string `json:"summaryType"`
	Level       string `json:"level"`
}

func (a *AdService) SummaryQuery(req SummaryQueryRequest, accessToken, AdvertiserId string) (map[string]interface{}, error) {
	return summaryQuery(a.c.Host, req, accessToken, AdvertiserId)
}

func summaryQuery(host string, req SummaryQueryRequest, accessToken, AdvertiserId string) (map[string]interface{}, error) {
	ms := time.Now().UnixNano() / 1e6
	qid := QidWithUnixTime()

	bts, _ := json.Marshal(req)
	nonce := fmt.Sprintf("%x", md5.Sum([]byte(qid)))
	url := fmt.Sprintf(buildMarketURL(host, summaryQueryURLFormat), accessToken, ms, nonce, AdvertiserId)
	resp, err := http.Post(url, "Content-Type: application/json", bytes.NewBuffer(bts))
	if err != nil {
		return nil, err
	}

	result, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(result))

	var response map[string]interface{}
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
