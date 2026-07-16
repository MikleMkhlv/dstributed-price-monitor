package repository

import (
	"context"
	"dstributed-price-monitor/config"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var KEY_PREFFIX string

type Redis struct {
	RClient   *redis.Client
	keyPrefix string
	ttl       time.Duration
}

func NewRedis(ctx context.Context, cfg config.Config) *Redis {
	rcl := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Username: cfg.Redis.Login,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.NumberDB,
	})

	_, err := rcl.Ping(ctx).Result()
	if err != nil {
		panic(fmt.Sprintf("failed ping redis: %v", err))
	}

	log.Print("repository.NewRedis: connection with redis is successful")
	KEY_PREFFIX = cfg.Redis.Db.KeyPrefix
	return &Redis{
		RClient: rcl,
		ttl:     cfg.Redis.Db.Ttl,
	}
}

func key(id string) string {
	return fmt.Sprintf("%s:%s", KEY_PREFFIX, id)
}

func (r *Redis) Close() {
	r.RClient.Close()
}

func (r *Redis) Put(ctx context.Context, id string, data []byte) error {
	key := key(id)
	err := r.RClient.Set(ctx, key, data, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("repository.Redis.Put:{%v} failed to save data: %v", ctx.Value("operId"), err)
	}
	log.Printf("repository.Redis.Put:{%v} put data in redis is successful: key=%s", ctx.Value("operId"), key)
	return nil
}

func (r *Redis) Get(ctx context.Context, id string) ([]byte, error) {
	key := key(id)
	data, err := r.RClient.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("repository.Redis.Get:{%v} data not found in redis by key %s", ctx.Value("operId"), key)
		}
		return nil, fmt.Errorf("repository.Redis.Get:{%v} failed get data from redis by key %s", ctx.Value("operId"), key)
	}
	log.Printf("repository.Redis.Get:{%v} get data in redis is successful: key=%s", ctx.Value("operId"), key)
	return data, nil
}

func (r *Redis) Delete(ctx context.Context, id string) error {
	key := key(id)
	isDel, err := r.RClient.Del(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("repository.Redis.Delete:{%v} failed delete data in redis by key %s", ctx.Value("operId"), key)
	}
	if isDel == 0 {
		return fmt.Errorf("repository.Redis.Delete:{%v} data is not found in redis by key %s", ctx.Value("operId"), key)
	}
	log.Printf("repository.Redis.Delete:{%v} delete data in redis is successful: key=%s", ctx.Value("operId"), key)
	return nil
}
