package tgfsm

import (
	"time"

	"go.uber.org/zap"
)

// Option is a function that modifies the bot configuration
type Option func(*Bot)

// WithExpiration sets the expiration time for user states
func WithExpiration(expiration time.Duration) Option {
	return func(b *Bot) {
		b.expiration = expiration
	}
}

// WithCleanupInterval sets the cleanup interval for user states
func WithCleanupInterval(cleanupInterval time.Duration) Option {
	return func(b *Bot) {
		b.cleanupInterval = cleanupInterval
	}
}

// WithStates sets the states for the bot
func WithStates(states map[string]State) Option {
	return func(b *Bot) {
		b.states = states
	}
}

// WithLogger sets the logger for the bot
func WithLogger(logger *zap.Logger) Option {
	return func(b *Bot) {
		b.logger = logger
	}
}

// WithUpdateHandler sets the update handler for the bot
func WithUpdateHandler(updateHandler HandlerFunc) Option {
	return func(b *Bot) {
		b.updateHandler = updateHandler
	}
}

// WithPrivateOnly sets the bot to only accept messages from private chats
func WithPrivateOnly(privateOnly bool) Option {
	return func(b *Bot) {
		b.privateOnly = privateOnly
	}
}

// WithBlacklistedChats sets the list of blacklisted chat IDs
func WithBlacklistedChats(chatIDs []int64) Option {
	return func(b *Bot) {
		b.blacklistedChats = make(map[int64]bool)
		for _, chatID := range chatIDs {
			b.blacklistedChats[chatID] = true
		}
	}
}

// WithAutoDelete enables auto-deletion of last message before sending new one
// Requires LastMessageCache to be set via WithLastMessageCache
func WithAutoDelete(enabled bool) Option {
	return func(b *Bot) {
		b.autoDeleteEnabled = enabled
	}
}

// WithLastMessageCache sets the cache implementation for last message IDs
// Used for auto-deletion feature
func WithLastMessageCache(cache LastMessageCache) Option {
	return func(b *Bot) {
		b.lastMessageCache = cache
	}
}
