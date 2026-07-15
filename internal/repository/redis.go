package repository

import (
	"context"
	"dstributed-price-monitor/config"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	RClient *redis.Client
}

func NewRedis(ctx context.Context, cfg config.Config) *Redis {
	rcl := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.NumberDB,
	})

	_, err := rcl.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("failed ping redis: %v", err))
	}

	log.Print("repository.NewRedis: connection with redis is successful")
	return &Redis{
		RClient: rcl,
	}
}

func (r *Redis) Close() {
	r.RClient.Close()
}
