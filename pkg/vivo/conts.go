package vivo

const (
	VIVO_MARKET_HOST       = "https://marketing-api.vivo.com.cn/openapi"
	GetVivoTokenUrl        = VIVO_MARKET_HOST + "/v1/oauth2/token?client_id=%s&client_secret=%s&grant_type=code&code=%s"
	RefreshVivoTokenUrl    = VIVO_MARKET_HOST + "/v1/oauth2/refreshToken?client_id=%s&client_secret=%s&refresh_token=%s"
	VivoAdvertiserQueryUrl = VIVO_MARKET_HOST + "/v1/account/fetch?access_token=%s&timestamp=%d&nonce=%s"
	VivoCallbackUrlV2      = VIVO_MARKET_HOST + "/v2/advertiser/behavior/upload?access_token=%s&timestamp=%d&nonce=%s"
)
