package redis

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/bondzai/logger/internal/util"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient() *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     util.GetEnv("REDIS_HOST", ""),
			Password: util.GetEnv("REDIS_PASSWORD", ""),
			DB:       getIntEnv("REDIS_DB", 0),
		}),
	}
}

func (r *RedisClient) SetData(cacheKey string, data interface{}, timeExpired time.Duration) error {
	jsonValue, err := json.Marshal(data)
	if err != nil {
		log.Printf("Error marshaling data to JSON: %v\n", err)
		return err
	}

	err = r.client.Set(context.TODO(), cacheKey, jsonValue, timeExpired).Err()
	if err != nil {
		log.Printf("Error storing data in Redis: %v\n", err)
		return err
	}

	return nil
}

func (r *RedisClient) GetData(cacheKey string, v interface{}) error {
	result, err := r.client.Get(context.TODO(), cacheKey).Result()
	if err != nil {
		log.Printf("Error retrieving data from Redis: %v\n", err)
		return err
	}

	err = json.Unmarshal([]byte(result), &v)
	if err != nil {
		log.Printf("Error unmarshaling JSON data: %v\n", err)
		return err
	}

	return nil
}

func getIntEnv(key string, defaultValue int) int {
	value, err := strconv.Atoi(util.GetEnv(key, strconv.Itoa(defaultValue)))
	if err != nil {
		log.Printf("Error converting environment variable %s to integer: %v\n", key, err)
		return defaultValue
	}
	return value
}
