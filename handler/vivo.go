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
			err := click(c.Request.Context(), logicSvc, channel, item)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "操作成功"})
	}
}
