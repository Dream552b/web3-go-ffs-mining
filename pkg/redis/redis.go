package redis

import (
	"context"
	"fss-mining/pkg/config"
	"fss-mining/pkg/logger"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var Client *goredis.Client

func Init(cfg config.RedisConfig) error {
	Client = goredis.NewClient(&goredis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := Client.Ping(ctx).Err(); err != nil {
		return err
	}

	logger.Info("Redis 连接成功", zap.String("addr", cfg.Addr()))
	return nil
}

func Get() *goredis.Client {
	return Client
}

// SetNX 用于幂等控制（防重复提交）
func SetNX(ctx context.Context, key string, value any, expiration time.Duration) (bool, error) {
	return Client.SetNX(ctx, key, value, expiration).Result()
}

// Del 删除 key
func Del(ctx context.Context, keys ...string) error {
	return Client.Del(ctx, keys...).Err()
}
