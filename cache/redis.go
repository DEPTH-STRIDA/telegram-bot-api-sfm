// Package cache provides implementations of LastMessageCache interface
// for auto-deletion feature.
//
// To use Redis implementation, install the dependency:
//
//	go get github.com/redis/go-redis/v9
package cache

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	tgfsm "tgfsm"

	"github.com/redis/go-redis/v9"
)

// RedisLastMessage implements LastMessageCache using Redis
type RedisLastMessage struct {
	client    *redis.Client
	ctx       context.Context
	keyPrefix string
}

// Ensure RedisLastMessage implements LastMessageCache interface
var _ tgfsm.LastMessageCache = (*RedisLastMessage)(nil)

// NewRedisLastMessage creates a new cache implementation using Redis
// client - Redis client
// keyPrefix - prefix for keys (default: "tgfsm:last_msg:")
func NewRedisLastMessage(client *redis.Client, keyPrefix string) *RedisLastMessage {
	if keyPrefix == "" {
		keyPrefix = "tgfsm:last_msg:"
	}
	return &RedisLastMessage{
		client:    client,
		ctx:       context.Background(),
		keyPrefix: keyPrefix,
	}
}

// GetLastMessageInfo returns information about the last message for a chat
func (c *RedisLastMessage) GetLastMessageInfo(chatID int64) (*tgfsm.LastMessageInfo, error) {
	key := c.keyPrefix + strconv.FormatInt(chatID, 10)
	value, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var info tgfsm.LastMessageInfo
	if err := json.Unmarshal([]byte(value), &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// SetLastMessageInfo saves information about the last message for a chat
// Uses default TTL of 30 days
func (c *RedisLastMessage) SetLastMessageInfo(chatID int64, info *tgfsm.LastMessageInfo) error {
	key := c.keyPrefix + strconv.FormatInt(chatID, 10)
	value, err := json.Marshal(info)
	if err != nil {
		return err
	}
	// Set TTL to 30 days
	return c.client.Set(c.ctx, key, value, 30*24*time.Hour).Err()
}

// DeleteLastMessageInfo deletes information about the last message for a chat
func (c *RedisLastMessage) DeleteLastMessageInfo(chatID int64) error {
	key := c.keyPrefix + strconv.FormatInt(chatID, 10)
	return c.client.Del(c.ctx, key).Err()
}
