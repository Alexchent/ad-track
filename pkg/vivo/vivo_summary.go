package vivo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	LevelACCOUNT  = "ACCOUNT"
	LevelCREATIVE = "CREATIVE"
	LevelCAMPAIGN = "CAMPAIGN"
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

	bts, _ := json.Marshal(req)
	nonce := MakeNonce()
	url := fmt.Sprintf(buildMarketURL(host, summaryQueryURLFormat), accessToken, ms, nonce, AdvertiserId)
	resp, err := http.Post(url, "Content-Type: application/json", bytes.NewBuffer(bts))
	if err != nil {
		return nil, err
	}

	result, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(result))

	var response map[string]interface{}
	err = json.Unmarshal(result, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
