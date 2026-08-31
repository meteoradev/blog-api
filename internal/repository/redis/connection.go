package redis

import (
	"context"
	"strconv"

	"github.com/meteoradev/BlogApi/config"
	"github.com/redis/go-redis/v9"
)

func RedisConnect(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost + ":" + strconv.Itoa(cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDatabase,
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return rdb, nil
}
