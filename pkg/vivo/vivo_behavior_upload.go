package vivo

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

var (
	VIVO_MARKET_HOST       = "https://marketing-api.vivo.com.cn/openapi"
	GetVivoTokenUrl        = VIVO_MARKET_HOST + "/v1/oauth2/token?client_id=%s&client_secret=%s&grant_type=code&code=%s"
	RefreshVivoTokenUrl    = VIVO_MARKET_HOST + "/v1/oauth2/refreshToken?client_id=%s&client_secret=%s&refresh_token=%s"
	VivoAdvertiserQueryUrl = VIVO_MARKET_HOST + "/v1/account/fetch?access_token=%s&timestamp=%d&nonce=%s"
	VivoCallbackUrlV2      = VIVO_MARKET_HOST + "/v2/advertiser/behavior/upload?access_token=%s&timestamp=%d&nonce=%s"

	// token刷新提前量：5分钟（毫秒），在token过期前5分钟自动刷新
	tokenRefreshBuffer int64 = 300000

	letterBytes   = "abcdef6789abcdefABCDEF67890123456789abcdef67890123ABCDEFabcdef"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

type BehaviorRequest struct {
	SrcType string     `json:"srcType,omitempty"`
	PkgName string     `json:"pkgName,omitempty"`
	SrcId   string     `json:"srcId,omitempty"` // 源ID
	Data    []DataList `json:"dataList,omitempty"`
}

type DataList struct {
	UserIdType string `json:"userIdType,omitempty"`
	UserId     string `json:"userId,omitempty"`
	CvType     string `json:"cvType,omitempty"`
	CvTime     int64  `json:"cvTime,omitempty"`
	ClickId    string `json:"clickId"` // vivo渠道事件（dataFrom=1）必传
}

type BehaviorResponse struct {
	Code    int64  `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data,omitempty"`
}

func BehaviorUpload(req *BehaviorRequest, accessToken string) (*BehaviorResponse, error) {
	ms := time.Now().UnixNano() / 1e6
	qid := QidWithUnixTime()
	nonce := fmt.Sprintf("%x", md5.Sum([]byte(qid)))
	url := fmt.Sprintf(VivoCallbackUrlV2, accessToken, ms, nonce)

	bts, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bts))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := &BehaviorResponse{}
	err = json.Unmarshal(result, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// String generate random string by length
func String(n int) string {
	src := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & int64(letterIdxMask)); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

// QidWithUnixTime 根据时间戳随机生成 16 位 1id
func QidWithUnixTime() (qid string) {
	now := time.Now().UnixNano() / 1000 // 16位纳秒
	nowStr := strconv.FormatInt(now, 10)
	withoutYear := nowStr[3:] // 去掉年开头的数字
	remainLen := 16 - len(withoutYear)
	ranStr := String(remainLen)
	return reverseString(withoutYear + ranStr)
}

func reverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}
