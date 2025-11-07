// Package cache provides implementations of LastMessageCache interface
// for auto-deletion feature.
package cache

import (
	"strconv"
	"sync"
	"time"

	tgfsm "tgfsm"

	gocache "github.com/patrickmn/go-cache"
)

// GoCacheLastMessage implements LastMessageCache using go-cache
type GoCacheLastMessage struct {
	cache *gocache.Cache
	mu    sync.RWMutex
}

// Ensure GoCacheLastMessage implements LastMessageCache interface
var _ tgfsm.LastMessageCache = (*GoCacheLastMessage)(nil)

// NewGoCacheLastMessage creates a new cache implementation using go-cache
// expiration - lifetime of entries in cache
// cleanupInterval - interval for cleaning up expired entries
func NewGoCacheLastMessage(expiration, cleanupInterval time.Duration) *GoCacheLastMessage {
	return &GoCacheLastMessage{
		cache: gocache.New(expiration, cleanupInterval),
	}
}

// GetLastMessageInfo returns information about the last message for a chat
func (c *GoCacheLastMessage) GetLastMessageInfo(chatID int64) (*tgfsm.LastMessageInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := strconv.FormatInt(chatID, 10)
	value, found := c.cache.Get(key)
	if !found {
		return nil, nil
	}

	info, ok := value.(*tgfsm.LastMessageInfo)
	if !ok {
		return nil, nil
	}

	return info, nil
}

// SetLastMessageInfo saves information about the last message for a chat
func (c *GoCacheLastMessage) SetLastMessageInfo(chatID int64, info *tgfsm.LastMessageInfo) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := strconv.FormatInt(chatID, 10)
	c.cache.Set(key, info, gocache.DefaultExpiration)
	return nil
}

// DeleteLastMessageInfo deletes information about the last message for a chat
func (c *GoCacheLastMessage) DeleteLastMessageInfo(chatID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := strconv.FormatInt(chatID, 10)
	c.cache.Delete(key)
	return nil
}
