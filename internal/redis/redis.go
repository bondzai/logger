package redis

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/bondzai/logger/internal/util"

	"github.com/redis/go-redis/v9"
)

var (
	RedisHost     = util.GetEnv("REDIS_HOST", "")
	RedisPassword = util.GetEnv("REDIS_PASSWORD", "")
	RedisDB, _    = strconv.Atoi(util.GetEnv("REDIS_DB", ""))
)

var (
	client *redis.Client
	once   sync.Once
)

func Initialize() *redis.Client {
	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Addr:     RedisHost,
			Password: RedisPassword,
			DB:       RedisDB,
		})
	})

	return client
}

func GetClient() *redis.Client {
	if client == nil {
		return Initialize()
	}

	return client
}

func SetData(cacheKey string, tasks interface{}) error {
	client := GetClient()

	jsonValue, err := json.Marshal(tasks)
	if err != nil {
		log.Printf("Error marshaling tasks to JSON: %v\n", err)
		return err
	}

	err = client.Set(context.TODO(), cacheKey, jsonValue, 24*time.Hour).Err()
	if err != nil {
		log.Printf("Error storing data in Redis: %v\n", err)
		return err
	}

	return nil
}

func GetData(cacheKey string, v interface{}) error {
	client := GetClient()

	result, err := client.Get(context.TODO(), cacheKey).Result()
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
