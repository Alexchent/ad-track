package vivo

import "strings"

const (
	DefaultVivoMarketHost = "https://marketing-api.vivo.com.cn"

	getVivoTokenURLFormat        = "/openapi/v1/oauth2/token?client_id=%s&client_secret=%s&grant_type=code&code=%s"
	refreshVivoTokenURLFormat    = "/openapi/v1/oauth2/refreshToken?client_id=%s&client_secret=%s&refresh_token=%s"
	vivoAdvertiserQueryURLFormat = "/openapi/v1/account/fetch?access_token=%s&timestamp=%d&nonce=%s"
	vivoCallbackURLV2Format      = "/openapi/v2/advertiser/behavior/upload?access_token=%s&timestamp=%d&nonce=%s"
	summaryQueryURLFormat        = "/openapi/v1/adstatement/summary/query?access_token=%s&timestamp=%d&nonce=%s&advertiser_id=%s"
)

func normalizeMarketHost(host string) string {
	if host == "" {
		host = DefaultVivoMarketHost
	}
	host = strings.TrimRight(host, "/")
	return host
}

func buildMarketURL(host, path string) string {
	return normalizeMarketHost(host) + path
}
