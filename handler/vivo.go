package handler

import (
	"fmt"
	"net/http"

	"github.com/Alexchent/ad-track/logic"
	"github.com/Alexchent/ad-track/svc"
	"github.com/gin-gonic/gin"
)

// GetAuthorizationCode vivo认证获取授权码
func GetAuthorizationCode(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		code := c.Query("code")
		if code == "" {
			c.String(http.StatusUnauthorized, "authorization code not provided")
		}
		// 获取点击参数
		adSvc := svcCtx.AdService
		response, err := adSvc.GetAccessToken(code)
		if err != nil {
			msg := fmt.Sprintf("get vivo access token failed %s", err.Error())
			c.String(http.StatusOK, msg)
			return
		}
		err = adSvc.SaveAccessToken(c.Request.Context(), response.Data)
		if err != nil {
			msg := fmt.Sprintf("vivo save access token failed %s", err.Error())
			c.String(http.StatusOK, msg)
			return
		}
		c.String(http.StatusOK, "ok")
	}
}

func ProcessClick(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// todo 获取点击参数, 兼容get和post

		// todo redis 保存点击数据，用 oaid 做key, 有效期7天

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
	}
}

// ProcessVIVOClick oppo服务端点击监测 https://ad.vivo.com.cn/help?id=352
func ProcessVIVOClick(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取点击参数
		channel := c.Query("channel")

		var body []map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
			return
		}

		logicSvc := logic.NewClick(svcCtx.Config)
		for _, item := range body {
			item[logic.Channel] = channel

			oaid, _ := item[logic.Oaid].(string)
			imei, _ := item[logic.Imei].(string)
			if oaid == "" && imei == "" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "oaid and imei are empty"})
				return
			}

			if oaid != "" {
				if err := logicSvc.SaveData(c.Request.Context(), oaid, item); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
					return
				}
			}

			if imei != "" {
				if err := logicSvc.SaveData(c.Request.Context(), imei, item); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
					return
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
	}
}

func AttributeReport(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 请求参数中需要携带 设备信息 OAID
		// 通过 OAID 匹配点击数据，通过点击数据中存储的 channel 判断要归因的媒体
		// 向对应的媒体上报
		// api := logic.NewVivoApi(c.Request.Context(), svcCtx.AdService, svcCtx.Config.VIVO)
	}
}
