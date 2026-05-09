package vivo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Alexchent/ad-track/pkg/uuid"
)

var (
	// token刷新提前量：5分钟（毫秒），在token过期前5分钟自动刷新
	tokenRefreshBuffer int64 = 300000
)

type BehaviorRequest struct {
	SrcType  string     `json:"srcType,omitempty"`
	PkgName  string     `json:"pkgName,omitempty"`
	SrcId    string     `json:"srcId,omitempty"` // 源ID
	DataForm string     `json:"dataForm,omitempty" default:"1"`
	Data     []DataList `json:"dataList,omitempty"`
}

type DataList struct {
	UserIdType string `json:"userIdType,omitempty"` // 用户标识类型, 枚举值 IMEI/IMEI_MD5/OAID/OAID_MD5/OTHER/OPENID/INSTALL_REFERRER
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

func (a *AdService) BehaviorUpload(req *BehaviorRequest, accessToken string) (*BehaviorResponse, error) {
	return behaviorUpload(a.c.Host, req, accessToken)
}

func behaviorUpload(host string, req *BehaviorRequest, accessToken string) (*BehaviorResponse, error) {
	ms := time.Now().UnixNano() / 1e6
	nonce := MakeNonce()
	url := fmt.Sprintf(buildMarketURL(host, vivoCallbackURLV2Format), accessToken, ms, nonce)

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

func MakeNonce() string {
	return uuid.MakeNonce()
}
