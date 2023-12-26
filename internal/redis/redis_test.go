package redis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedisReadWrite(t *testing.T) {
	client := Initialize()

	ctx := context.Background()

	err := client.Set(ctx, "test_key", "test_value", 0).Err()
	if err != nil {
		t.Fatalf("Failed to set key: %s", err)
	}

	val, err := client.Get(ctx, "test_key").Result()
	if err != nil {
		t.Fatalf("Failed to get key: %s", err)
	}

	assert.Equal(t, "test_value", val, "Value from Redis should match the set value")
}
