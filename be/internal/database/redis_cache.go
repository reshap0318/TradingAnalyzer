package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheClient wraps Redis client with cache-specific operations
type CacheClient struct {
	client *redis.Client
}

// NewCacheClient creates a new cache client wrapper
func NewCacheClient(client *redis.Client) *CacheClient {
	return &CacheClient{
		client: client,
	}
}

// GetClient returns the underlying Redis client
func (c *CacheClient) GetClient() *redis.Client {
	return c.client
}

// Ping checks Redis connection
func (c *CacheClient) Ping() error {
	ctx := context.Background()
	return c.client.Ping(ctx).Err()
}

// Set stores a string value in Redis with expiration
func (c *CacheClient) Set(ctx context.Context, key, value string, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a string value from Redis
func (c *CacheClient) Get(ctx context.Context, key string) (string, error) {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key not found
	}
	return val, err
}

// Delete removes a key from Redis
func (c *CacheClient) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

// Exists checks if keys exist in Redis
func (c *CacheClient) Exists(ctx context.Context, keys ...string) (bool, error) {
	count, err := c.client.Exists(ctx, keys...).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SetJSON stores a JSON-encoded value in Redis with expiration
func (c *CacheClient) SetJSON(ctx context.Context, key string, data interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return c.client.Set(ctx, key, string(jsonData), expiration).Err()
}

// GetJSON retrieves and decodes a JSON value from Redis
func (c *CacheClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil // Key not found
	}
	if err != nil {
		return err
	}

	if val == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// SetWithTTL stores a value with time-to-live
func (c *CacheClient) SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

// GetTTL gets the remaining time-to-live of a key
func (c *CacheClient) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}

// Expire sets a timeout on a key
func (c *CacheClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.client.Expire(ctx, key, expiration).Err()
}

// FlushDB removes all keys from the current database
func (c *CacheClient) FlushDB(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Keys returns all keys matching a pattern
func (c *CacheClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	return c.client.Keys(ctx, pattern).Result()
}

// SetNX sets a key only if it does not exist
func (c *CacheClient) SetNX(ctx context.Context, key, value string, expiration time.Duration) (bool, error) {
	return c.client.SetNX(ctx, key, value, expiration).Result()
}

// Incr increments the integer value of a key
func (c *CacheClient) Incr(ctx context.Context, key string) (int64, error) {
	return c.client.Incr(ctx, key).Result()
}

// Decr decrements the integer value of a key
func (c *CacheClient) Decr(ctx context.Context, key string) (int64, error) {
	return c.client.Decr(ctx, key).Result()
}

// SAdd adds members to a set
func (c *CacheClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SAdd(ctx, key, members...).Err()
}

// SRem removes members from a set
func (c *CacheClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	return c.client.SRem(ctx, key, members...).Err()
}

// SMembers returns all members of a set
func (c *CacheClient) SMembers(ctx context.Context, key string) ([]string, error) {
	return c.client.SMembers(ctx, key).Result()
}

// SIsMember checks if a member exists in a set
func (c *CacheClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return c.client.SIsMember(ctx, key, member).Result()
}

// LPush pushes values to a list
func (c *CacheClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	return c.client.LPush(ctx, key, values...).Err()
}

// RPop pops the last element from a list
func (c *CacheClient) RPop(ctx context.Context, key string) (string, error) {
	return c.client.RPop(ctx, key).Result()
}

// LRange returns a range of elements from a list
func (c *CacheClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return c.client.LRange(ctx, key, start, stop).Result()
}

// LLen returns the length of a list
func (c *CacheClient) LLen(ctx context.Context, key string) (int64, error) {
	return c.client.LLen(ctx, key).Result()
}

// HSet sets fields in a hash
func (c *CacheClient) HSet(ctx context.Context, key string, values ...interface{}) error {
	return c.client.HSet(ctx, key, values...).Err()
}

// HGet gets a field from a hash
func (c *CacheClient) HGet(ctx context.Context, key, field string) (string, error) {
	return c.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all fields and values from a hash
func (c *CacheClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return c.client.HGetAll(ctx, key).Result()
}

// HDel deletes fields from a hash
func (c *CacheClient) HDel(ctx context.Context, key string, fields ...string) error {
	return c.client.HDel(ctx, key, fields...).Err()
}
