package lib

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type RedisClient struct {
	client *redis.Client
}

func (redisClient RedisClient) Set(key string, value interface{}, expire time.Duration) error {
	return redisClient.client.Set(context.Background(), key, value, expire).Err()
}

func (redisClient RedisClient) Get(key string) (string, error) {
	return redisClient.client.Get(context.Background(), key).Result()
}

func (redisClient RedisClient) Close() error {
	return redisClient.client.Close()
}

func (redisClient RedisClient) Ping() (string, error) {
	str, err := redisClient.client.Ping(context.Background()).Result()
	return str, err
}

func NewRedisClient() *RedisClient {

	redisOptions := &redis.Options{
		Addr:     environment[redisAddress],
		Password: environment[redisSecret],
		DB:       0,
	}

	client := redis.NewClient(redisOptions)

	redisClient := &RedisClient{
		client: client,
	}
	log.Print(redisOptions.Addr, redisOptions.Password)
	return redisClient
}
