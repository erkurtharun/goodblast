package redisclient

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"goodblast/config"
	"goodblast/pkg/log"
	"time"
)

type Client struct {
	Client *redis.Client
}

func NewClient(config *appconfig.Config) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.RedisHost, config.RedisPort),
		Password: config.RedisPassword,
		DB:       config.RedisDB,
		Protocol: config.RedisConnectionProtocol,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	log.GetLogger().Info("Connected to Redis successfully")

	return &Client{Client: rdb}
}
