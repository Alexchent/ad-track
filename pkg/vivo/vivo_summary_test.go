package vivo

import (
	"os"
	"testing"
	"time"
)

var token string
var advertiserId string
var adService *AdService

func init() {
	token = os.Getenv("VIVO_TOKEN")
	if token == "" {
		panic("VIVO_TOKEN environment variable is not set")
	}
	adService = &AdService{
		c: &Config{
			Host: "https://marketing-api.vivo.com.cn",
		},
	}
}

func TestQueryAdvertiser(t *testing.T) {
	advertiser, err := adService.QueryAdvertiser(token)
	if err != nil {
		return
	}
	t.Log(advertiser)
	advertiserId = advertiser.Data.UUID
}

func TestSummaryQuery(t *testing.T) {
	if advertiserId == "" {
		//t.Skip("advertiserId is empty, skipping (run TestQueryAdvertiser first)")
		advertiser, err := adService.QueryAdvertiser(token)
		if err != nil {
			return
		}
		t.Log(advertiser)
		advertiserId = advertiser.Data.UUID
	}

	t.Run("t1", func(t *testing.T) {
		res := SummaryQueryRequest{
			StartDate:   time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			EndDate:     time.Now().Format("2006-01-02"),
			PageIndex:   0,
			PageSize:    100,
			SummaryType: "DAY",
			Level:       LevelACCOUNT,
		}
		_, err := adService.SummaryQuery(res, token, advertiserId)
		if err != nil {
			t.Fatalf("query fail %s", err.Error())
		}
	})
}
