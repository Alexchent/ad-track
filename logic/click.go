package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Alexchent/ad-track/config"
	"github.com/go-redis/redis/v8"
)

const (
	ClickKeyPrefix = "click:"
	clickTTL       = 30 * 24 * time.Hour
)

type Click struct {
	cache *redis.Client
}

func NewClick(c config.Config) *Click {
	return &Click{
		cache: redis.NewClient(&redis.Options{Addr: c.Redis.Addr, Password: c.Redis.Password, DB: c.Redis.Db}),
	}
}

// SaveData 保存监测数据
//
//	deviceKey 设备标识：
//		IMEI：15-17位，明文
//		IMEI_MD5 ：32位，加密
//		OAID：64位，明文
//		OAID_MD5：32位   加密
func (c *Click) SaveData(ctx context.Context, deviceKey string, data map[string]interface{}) error {
	if deviceKey == "" {
		return fmt.Errorf("key is empty")
	}

	key := ClickKeyPrefix + deviceKey
	values := make(map[string]interface{}, len(data))
	for k, v := range data {
		values[k] = stringifyRedisValue(v)
	}

	if err := c.cache.HSet(ctx, key, values).Err(); err != nil {
		return err
	}
	return c.cache.Expire(ctx, key, clickTTL).Err()
}

func (c *Click) GetData(ctx context.Context, deviceKey string) (map[string]string, error) {
	if deviceKey == "" {
		return nil, fmt.Errorf("key is empty")
	}

	data, err := c.cache.HGetAll(ctx, ClickKeyPrefix+deviceKey).Result()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func stringifyRedisValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case fmt.Stringer:
		return val.String()
	default:
		b, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprint(val)
		}
		return string(b)
	}
}
