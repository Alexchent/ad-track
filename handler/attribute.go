package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/Alexchent/ad-track/logic"
	"github.com/Alexchent/ad-track/svc"
	"github.com/gin-gonic/gin"
)

func AttributeReport(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求参数中需要携带设备标识信息 OAID 或 IMEI，通过设备标识匹配点击数据。
		oaid := strings.TrimSpace(c.Query(logic.Oaid))
		imei := strings.TrimSpace(c.Query(logic.Imei))
		if oaid == "" && imei == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "oaid and imei are empty"})
			return
		}

		logicSvc := logic.NewClick(svcCtx.Config)
		clickData, _, err := findClickData(c.Request.Context(), logicSvc, oaid, imei)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
			return
		}
		if len(clickData) == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "click data not found"})
			return
		}

		channel := clickData[logic.Channel]
		if !strings.Contains(strings.ToLower(channel), "vivo") {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "not vivo channel"})
			return
		}

		clickData[logic.AppUid] = c.Query("user_id")
		clickData[logic.PkgName] = c.Query("package_name")

		var api logic.Attribute
		if strings.Contains(strings.ToLower(channel), "vivo") {
			api = logic.NewVivoApi(c.Request.Context(), svcCtx.AdService, svcCtx.Config.VIVO)
		}
		if err := api.Active(clickData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
	}
}

func findClickData(ctx context.Context, logicSvc *logic.Click, deviceIds ...string) (map[string]string, string, error) {
	for _, key := range deviceIds {
		data, err := logicSvc.GetData(ctx, key)
		if err != nil {
			return nil, "", err
		}
		if len(data) > 0 {
			return data, key, nil
		}
	}
	return nil, "", nil
}
