package vivo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

type Config struct {
	Host         string `json:"host"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type AdService struct {
	c           *Config
	AccessToken string `json:"access_token"`
	redisClient *redis.Client
}

func NewAdService(c *Config) *AdService {
	if c == nil {
		c = &Config{}
	}
	c.Host = normalizeMarketHost(c.Host)
	return &AdService{
		c: c,
	}
}

type VivoRepsonse struct {
	Code    int64  `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Data    AdvertiserToken
}

// AdvertiserToken 单个广告主的token信息（用于统一存储）
type AdvertiserToken struct {
	AccessToken      string `json:"access_token,omitempty"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	TokenDate        int64  `json:"token_date,omitempty"`         // 时间戳(毫秒),accessToken截止的有效日期
	RefreshTokenDate int64  `json:"refresh_token_date,omitempty"` // 时间戳(毫秒),refreshTokenDate截止的有效日期
}

func (a *AdService) GetAccessToken(code string) (*VivoRepsonse, error) {
	url := fmt.Sprintf(buildMarketURL(a.c.Host, getVivoTokenURLFormat), a.c.ClientId, a.c.ClientSecret, code)
	resp, err := http.Get(url)
	if err != nil {
		msg := "send http to get vivo token request fail"
		return nil, errors.New(msg)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := "read vivo token response fail," + err.Error()
		return nil, errors.New(msg)
	}

	response := &VivoRepsonse{}
	err = json.Unmarshal(data, response)
	if err != nil {
		msg := "json unmarshal  vivo struct fail," + err.Error()
		return nil, errors.New(msg)
	}

	if response.Code != 0 {
		msg := "vivo get toekn failed error message: " + response.Message
		return nil, errors.New(msg)
	}

	return response, nil
}

// SaveAccessToken 根据Authorization Code 生成token
func (a *AdService) SaveAccessToken(ctx context.Context, token AdvertiserToken) error {
	// 获取 广告主uuid
	advertiser, err := queryAdvertiser(a.c.Host, token.AccessToken)
	if err != nil {
		return err
	}
	advertiserId := advertiser.Data.UUID
	if err := a.saveTokenToRedis(ctx, advertiserId, &token); err != nil {
		return err
	}
	return nil
}

// getTokenRedisKey 获取vivo token存储的Redis key
// 每个广告主独立存储: key=vivo_token_{clientId}_{advertiserId}
func getTokenRedisKey(clientId string, advertiserId string) string {
	return fmt.Sprintf("vivo_token_%s_%s", clientId, advertiserId)
}

// saveTokenToRedis 保存单个广告主的token信息到Redis
// Redis结构: key=vivo_token_{clientId}_{advertiserId}, value=JSON(AdvertiserToken)
// 根据RefreshTokenDate设置key的独立过期时间
func (a *AdService) saveTokenToRedis(ctx context.Context, advertiserId string, tokenInfo *AdvertiserToken) error {
	clientId := a.c.ClientId
	key := getTokenRedisKey(clientId, advertiserId)
	jsonData, err := json.Marshal(tokenInfo)
	if err != nil {
		return err
	}

	var saveErr error
	for i := 0; i < 3; i++ {
		saveErr = a.redisClient.Set(ctx, key, string(jsonData), 0).Err()
		if saveErr == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if saveErr != nil {
		return saveErr
	}

	// 根据 RefreshTokenDate 设置key的独立过期时间
	if tokenInfo.RefreshTokenDate > 0 {
		nowMs := time.Now().UnixNano() / 1e6
		remainingMs := tokenInfo.RefreshTokenDate - nowMs
		if remainingMs > 0 {
			newTTL := time.Duration(remainingMs) * time.Millisecond
			a.redisClient.Expire(ctx, key, newTTL)
		}
	}

	return nil
}

// RefreshToken 使用refresh_token刷新access_token
func (a *AdService) RefreshToken(ctx context.Context, refreshToken string, advertiserId string) (*AdvertiserToken, error) {
	url := fmt.Sprintf(buildMarketURL(a.c.Host, refreshVivoTokenURLFormat), a.c.ClientId, a.c.ClientSecret, refreshToken)
	resp, err := http.Get(url)
	if err != nil {
		msg := "send http to refresh vivo token request fail"
		return nil, errors.New(msg)
	}
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		msg := "read vivo refresh token response fail," + err.Error()
		return nil, errors.New(msg)
	}

	response := &VivoRepsonse{}
	err = json.Unmarshal(data, response)

	if err != nil {
		msg := "json unmarshal vivo refresh token struct fail," + err.Error()
		return nil, errors.New(msg)
	}

	if response.Code != 0 {
		msg := "vivo refresh token failed error message " + response.Message
		return nil, errors.New(msg)
	}

	// 保存新的 token 到Redis统一存储
	tokenInfo := &response.Data
	if err := a.saveTokenToRedis(ctx, advertiserId, tokenInfo); err != nil {
		return tokenInfo, err
	}
	return tokenInfo, nil
}

// isTokenExpired 检查access_token是否即将过期或已过期
// tokenDate为accessToken截止的有效日期（毫秒时间戳）
// 当当前时间距离过期时间不足tokenRefreshBuffer时，认为需要刷新
func isTokenExpired(tokenDate int64) bool {
	if tokenDate == 0 {
		// 没有过期时间信息，不触发自动刷新
		return false
	}
	nowMs := time.Now().UnixNano() / 1e6
	// 当前时间 >= 过期时间 - 提前量，则需要刷新
	return nowMs >= tokenDate-tokenRefreshBuffer
}

// GetToken 获取指定广告主的access_token和refresh_token（带自动刷新）
// clientId: 开发者账号ID
// advertiserId: 广告主账号ID
func (a *AdService) GetToken(ctx context.Context, advertiserId string) (string, string) {
	clientId := a.c.ClientId
	key := getTokenRedisKey(clientId, advertiserId)
	for i := 0; i < 3; i++ {
		data, err := a.redisClient.Get(ctx, key).Result()
		if err == nil && data != "" {
			tokenInfo := &AdvertiserToken{}
			if jsonErr := json.Unmarshal([]byte(data), tokenInfo); jsonErr == nil {
				if tokenInfo.AccessToken == "" || tokenInfo.AccessToken == " " {
					return "", ""
				}
				// 检查access_token是否即将过期或已过期，自动刷新
				//if isTokenExpired(tokenInfo.TokenDate) && tokenInfo.RefreshToken != "" {
				//	newTokenInfo, refreshErr := RefreshToken(ctx, clientId, Secret, tokenInfo.RefreshToken, advertiserId)
				//	if refreshErr == nil {
				//		ctx.Warning(fmt.Sprintf("vivo auto refresh token for advertiser %s: %s", advertiserId, newTokenInfo.AccessToken))
				//		return newTokenInfo.AccessToken, newTokenInfo.RefreshToken
				//	}
				//	// 刷新失败，返回旧token（可能仍在短暂有效期内）
				//	ctx.Warning(fmt.Sprintf("vivo auto refresh token failed for advertiser %s: %v, using old token", advertiserId, refreshErr))
				//}
				return tokenInfo.AccessToken, tokenInfo.RefreshToken
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	return "", ""
}
