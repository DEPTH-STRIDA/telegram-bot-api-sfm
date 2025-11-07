package tgfsm

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	// Telegram API limits
	GlobalMessageLimit = 30 // Maximum messages per second to all chats
	ChatMessageLimit   = 1  // Maximum messages per second to one chat
	APILimit           = 30 // Maximum API requests per second
	BurstSize          = 5  // Burst allowance for rate limiter
)

// Limiter manages rate limiting for Telegram API requests
type Limiter struct {
	mu sync.RWMutex

	// Global rate limiter for all messages
	globalLimiter *rate.Limiter

	// Per-chat rate limiters
	chatLimiters map[int64]*rate.Limiter

	// API rate limiter
	apiLimiter *rate.Limiter
}

// NewLimiter creates a new rate limiter
func NewLimiter() *Limiter {
	return &Limiter{
		globalLimiter: rate.NewLimiter(rate.Every(time.Second/GlobalMessageLimit), BurstSize),
		chatLimiters:  make(map[int64]*rate.Limiter),
		apiLimiter:    rate.NewLimiter(rate.Every(time.Second/APILimit), BurstSize),
	}
}

// WaitForMessage waits for permission to send a message to a specific chat
func (l *Limiter) WaitForMessage(ctx context.Context, chatID int64) error {
	// Wait for global message limit
	if err := l.globalLimiter.Wait(ctx); err != nil {
		return err
	}

	// Wait for per-chat limit
	chatLimiter := l.getChatLimiter(chatID)
	return chatLimiter.Wait(ctx)
}

// WaitForAPI waits for permission to make an API request
func (l *Limiter) WaitForAPI(ctx context.Context) error {
	return l.apiLimiter.Wait(ctx)
}

// getChatLimiter returns or creates a rate limiter for a specific chat
func (l *Limiter) getChatLimiter(chatID int64) *rate.Limiter {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limiter, exists := l.chatLimiters[chatID]; exists {
		return limiter
	}

	// Create new limiter for this chat (1 message per second)
	limiter := rate.NewLimiter(rate.Every(time.Second/ChatMessageLimit), 1)
	l.chatLimiters[chatID] = limiter
	return limiter
}

// AllowMessage checks if a message can be sent without waiting
func (l *Limiter) AllowMessage(chatID int64) bool {
	// Check global limit
	if !l.globalLimiter.Allow() {
		return false
	}

	// Check per-chat limit
	chatLimiter := l.getChatLimiter(chatID)
	return chatLimiter.Allow()
}

// AllowAPI checks if an API request can be made without waiting
func (l *Limiter) AllowAPI() bool {
	return l.apiLimiter.Allow()
}
