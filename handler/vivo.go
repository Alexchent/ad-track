package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Alexchent/ad-track/pkg/vivo"
	"github.com/gin-gonic/gin"
)

const (
	cvTypeActivation = "ACTIVATION"  //自定义激活
	cvTypeRetention  = "RETENTION_1" //自定义次留
	cvTypeRetention2 = "RETENTION_2" //2日留存
	cvTypeRetention3 = "RETENTION_3" //3日留存
	cvTypeRetention4 = "RETENTION_4" //4日留存
	cvTypeRetention5 = "RETENTION_5" //5日留存
	cvTypeRetention6 = "RETENTION_6" //6日留存
	cvTypeRetention7 = "RETENTION_7" //7日留存

	userIdTypeOaid = "OAID" //行为数据上传 用户标识类型 oaid
	userIdTypeImei = "IMEI" //imei md5
)

const (
	Imei         = "imei"
	Oaid         = "oaid"
	Channel      = "channel"
	AppUid       = "app_uid"
	PkgName      = "pkgName"
	AdvertiserId = "advertiserId" // vivo投放广告主ID https://ad.vivo.com.cn/help?id=352
)

// GetAuthorizationCode vivo认证获取授权码
func GetAuthorizationCode(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.String(http.StatusUnauthorized, "authorization code not provided")
	}
	// 获取点击参数
	svc := vivo.NewAdService(&vivo.Config{
		ClientId:     "",
		ClientSecret: "",
	})
	response, err := svc.GetAccessToken(code)
	if err != nil {
		msg := fmt.Sprintf("get vivo access token failed %s", err.Error())
		c.String(http.StatusOK, msg)
		return
	}
	err = svc.SaveAccessToken(c.Request.Context(), response.Data)
	if err != nil {
		msg := fmt.Sprintf("vivo save access token failed %s", err.Error())
		c.String(http.StatusOK, msg)
		return
	}
	c.String(http.StatusOK, "ok")
}

// ProcessVIVOClick 接收oppo点击数据
func ProcessVIVOClick(c *gin.Context) {
	// 获取点击参数
	channel, _ := c.GetQuery("channel")

	var body []map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body = append(body, map[string]interface{}{"channel": channel})
	// todo 保存点击数据，归因时使用

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
}

type VivoApi struct {
	svc *vivo.AdService
	ctx context.Context
}

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

func (v *VivoApi) Active(data map[string]string) error {
	return v.callbackVivoBehavior(data, cvTypeActivation)
}

func (v *VivoApi) RetainNextDay(data map[string]string) error {
	return v.callbackVivoBehavior(data, cvTypeRetention)
}

func (v *VivoApi) Retain2(data map[string]string) error {
	return v.callbackVivoBehavior(data, cvTypeRetention2)
}

func (v *VivoApi) callbackVivoBehavior(d map[string]string, uType string) error {
	ctx := v.ctx
	if d[Oaid] == "" && d[Imei] == "" {
		return errors.New("device id is empty")
	}

	advertiserId, ok := d[AdvertiserId]
	if !ok || advertiserId == "" {
		return errors.New("advertiser id is empty")
	}
	userid := d[AppUid]

	accessToken, _ := v.svc.GetVivoToken(ctx, advertiserId)
	if accessToken == "" {
		return errors.New("vivo callback empty access token")
	}

	req := vivo.BehaviorRequest{
		SrcType: "APP",
		PkgName: d[PkgName],
		SrcId:   "", // todo 配置中获取
	}
	ms := time.Now().UnixNano() / 1e6
	obj := &vivo.DataList{
		UserIdType: userIdTypeOaid,
		UserId:     d[Oaid],
		ClickId:    d["ClickId"],
		CvType:     uType,
		CvTime:     ms,
	}
	req.Data = []vivo.DataList{*obj}

	bts, _ := json.Marshal(req)

	response, err := vivo.BehaviorUpload(&req, accessToken)
	if err != nil {
		return err
	}
	if response.Code != 0 {
		message := fmt.Sprintf("VIVO behaviorUpload fail: type-[%v], uid-[%v], params-[%s], code=%d, msg=%s", uType, userid, string(bts), response.Code, response.Message)
		return fmt.Errorf(message)
	}
	return nil
}
