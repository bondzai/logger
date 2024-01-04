package redis

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestRedisClient tests the functionality of the RedisClient.
func TestRedisClient(t *testing.T) {
	// Create a new instance of RedisClient
	redisClient := NewRedisClient()

	// Test SetData and GetData
	cacheKey := "testKey"
	testData := map[string]string{"key1": "value1", "key2": "value2"}

	err := redisClient.SetData(cacheKey, testData, 1*time.Second)
	assert.NoError(t, err, "SetData should not return an error")

	var resultData map[string]string
	err = redisClient.GetData(cacheKey, &resultData)
	assert.NoError(t, err, "GetData should not return an error")
	assert.Equal(t, testData, resultData, "Stored and retrieved data should be equal")

	// Test SetData with expiration
	cacheKeyWithExp := "testKeyWithExp"
	err = redisClient.SetData(cacheKeyWithExp, testData, 1*time.Second)
	assert.NoError(t, err, "SetData should not return an error")

	// Sleep for a while to let the data expire
	time.Sleep(2 * time.Second) // Adjust the sleep duration based on your expiration time

	err = redisClient.GetData(cacheKeyWithExp, &resultData)
	assert.Error(t, err, "GetData should return an error for expired data")
	assert.Contains(t, err.Error(), "redis: nil", "Error should indicate expired data")
}
