package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type Client struct {
	*redis.Client
	keyPrefix string
}

func NewClient(opt *redis.Options, keyPrefix string) *Client {
	return &Client{
		Client:    redis.NewClient(opt),
		keyPrefix: keyPrefix,
	}
}

func Wrap(client *redis.Client, keyPrefix string) *Client {
	return &Client{
		Client:    client,
		keyPrefix: keyPrefix,
	}
}

func (c *Client) HSet(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return c.Client.HSet(ctx, c.key(key), values...)
}

func (c *Client) HGetAll(ctx context.Context, key string) *redis.StringStringMapCmd {
	return c.Client.HGetAll(ctx, c.key(key))
}

func (c *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return c.Client.Set(ctx, c.key(key), value, expiration)
}

func (c *Client) Get(ctx context.Context, key string) *redis.StringCmd {
	return c.Client.Get(ctx, c.key(key))
}

func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) *redis.BoolCmd {
	return c.Client.Expire(ctx, c.key(key), expiration)
}

func (c *Client) key(key string) string {
	if c == nil {
		return key
	}
	return c.keyPrefix + key
}
