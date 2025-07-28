// internal/services/redis.go
package services

import (
	"context"
	"fmt"
	"log"
	"sync" // 导入 sync 包
	"time"

	"github.com/Hope-Cruiser-Psy-Volunteer-Alliance/HOMB/internal/config"
	"github.com/go-redis/redis/v8"
)

type RedisService interface {
	GetClient() *redis.Client
	Close() error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
}

type redisServiceImpl struct {
	client *redis.Client
}

var (
	redisSingleton RedisService
	redisOnce      sync.Once
)

func GetRedisService() (RedisService, error) {
	var initErr error
	redisOnce.Do(func() {
		redisConfig := config.Config.Redis

		client := redis.NewClient(&redis.Options{
			Addr:     redisConfig.Addr,
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, pingErr := client.Ping(ctx).Result()
		if pingErr != nil {
			initErr = fmt.Errorf("无法连接到 Redis: %w", pingErr)
			return
		}

		log.Println("Redis 连接成功！")

		redisSingleton = &redisServiceImpl{client: client}
	})

	return redisSingleton, initErr
}

func (s *redisServiceImpl) GetClient() *redis.Client {
	return s.client
}

func (s *redisServiceImpl) Close() error {
	if s.client != nil {
		log.Println("关闭 Redis 连接...")
		return s.client.Close()
	}
	return nil
}

func (s *redisServiceImpl) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return s.client.Set(ctx, key, value, expiration)
}

func (s *redisServiceImpl) Get(ctx context.Context, key string) *redis.StringCmd {
	return s.client.Get(ctx, key)
}

func (s *redisServiceImpl) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	return s.client.Del(ctx, keys...)
}
