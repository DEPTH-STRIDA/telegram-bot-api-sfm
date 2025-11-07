package tgfsm

// LastMessageInfo contains information about the last message
type LastMessageInfo struct {
	MessageID int
	Important bool // true if message is important (should not be deleted)
}

// LastMessageCache interface for caching last messages
// Stores last message ID and importance flag for each chat
type LastMessageCache interface {
	// GetLastMessageInfo returns information about the last message for a chat
	// Returns nil if message is not found
	GetLastMessageInfo(chatID int64) (*LastMessageInfo, error)

	// SetLastMessageInfo saves information about the last message for a chat
	SetLastMessageInfo(chatID int64, info *LastMessageInfo) error

	// DeleteLastMessageInfo deletes information about the last message for a chat
	DeleteLastMessageInfo(chatID int64) error
}
