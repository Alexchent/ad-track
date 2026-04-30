package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Alexchent/ad-track/config"
	"github.com/Alexchent/ad-track/pkg/vivo"
	"go.uber.org/zap"
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

	userIdTypeOaid    = "OAID"     // oaid 明文
	userIdTypeOaidMD5 = "OAID_MD5" // oaid md5
	userIdTypeImei    = "IMEI"     // imei 明文
	userIdTypeImeiMD5 = "IMEI_MD5" // imei md5
)

const (
	Imei         = "imei"
	Oaid         = "oaid"
	Androidid    = "androidid"
	Channel      = "channel"
	AppUid       = "app_uid"
	PkgName      = "pkgName"
	AdvertiserId = "advertiserId" // vivo投放广告主ID https://ad.vivo.com.cn/help?id=352
)

type VivoApi struct {
	svc     *vivo.AdService
	ctx     context.Context
	appList map[string]string
}

func NewVivoApi(ctx context.Context, svc *vivo.AdService, conf config.VIVOConfig) *VivoApi {
	return &VivoApi{
		ctx:     ctx,
		appList: conf.APP,
		svc:     svc,
	}
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

func (v *VivoApi) getSrcId(pkgName string) string {
	src, ok := v.appList[pkgName]
	if !ok {
		return ""
	}
	return src
}

func getDeviceID(d map[string]string) (string, string) {
	if d[Oaid] != "" {
		if len(d[Oaid]) == 32 {
			return userIdTypeOaidMD5, d[Oaid]
		}
		return userIdTypeOaid, d[Oaid]
	}

	if d[Imei] != "" {
		if len(d[Imei]) == 32 {
			return userIdTypeImeiMD5, d[Imei]
		}
		return userIdTypeImei, d[Imei]
	}

	return "", ""
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

	accessToken, _ := v.svc.GetToken(ctx, advertiserId)
	if accessToken == "" {
		slog.Info("VivoApi get access token failed", zap.String("advertiser_id", advertiserId))
		return fmt.Errorf("vivo callback advertiserId = %s empty accessToken", advertiserId)
	}

	srcId, ok := v.appList[d[PkgName]]
	if !ok {
		return fmt.Errorf("srcId not found for package %s ", d[PkgName])
	}

	req := vivo.BehaviorRequest{
		SrcType: "APP",
		PkgName: d[PkgName],
		SrcId:   srcId,
	}
	ms := time.Now().UnixNano() / 1e6
	userIdType, deviceId := getDeviceID(d)
	clickID := d["clickId"]
	if clickID == "" {
		clickID = d["ClickId"]
	}
	obj := &vivo.DataList{
		UserIdType: userIdType,
		UserId:     deviceId,
		ClickId:    clickID,
		CvType:     uType,
		CvTime:     ms,
	}
	req.Data = []vivo.DataList{*obj}

	bts, _ := json.Marshal(req)

	response, err := v.svc.BehaviorUpload(&req, accessToken)
	if err != nil {
		return err
	}
	if response.Code != 0 {
		slog.Error("VIVO behaviorUpload fail",
			slog.String("type", uType),
			slog.Int64("code", response.Code),
			slog.String("message", response.Message),
			slog.String("userId", userid),
			slog.String("clickId", clickID),
			slog.String("params", string(bts)))
		return fmt.Errorf("VIVO behaviorUpload fail code=%d message=%s", response.Code, response.Message)
	}
	return nil
}
