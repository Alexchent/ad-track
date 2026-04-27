package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Alexchent/ad-track/config"
	"github.com/Alexchent/ad-track/pkg/vivo"
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
