package vivo

import (
	"testing"
	"time"
)

func TestSummaryQuery(t *testing.T) {
	t.Run("t1", func(t *testing.T) {
		res := SummaryQueryRequest{
			StartDate: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			EndDate:   time.Now().Format("2006-01-02"),
			//StartDate:   "2025-08-01",
			//EndDate:     "2025-08-12",
			PageIndex:   0,
			PageSize:    100,
			SummaryType: "DAY",
			//Level:       "CREATIVE",
			//Level:       "CAMPAIGN",
			Level: "ACCOUNT",
		}
		_, err := SummaryQuery(res, "", "")
		if err != nil {
			t.Fatalf("query fail %s", err.Error())
		}
	})
}
