package tgfsm

import (
	"errors"
	"fmt"
)

var (
	// ErrStatesNil is returned when the states map is nil
	ErrStatesNil = fmt.Errorf("states map is nil")

	// ErrInvalidToken is returned when the bot token is empty or invalid
	ErrInvalidToken = fmt.Errorf("invalid bot token")

	// ErrNegativeExpiration is returned when expiration time is negative
	ErrNegativeExpiration = fmt.Errorf("expiration time cannot be negative")

	// ErrNegativeCleanup is returned when cleanup interval is negative
	ErrNegativeCleanup = fmt.Errorf("cleanup interval cannot be negative")

	// ErrTelegramInit is returned when Telegram API initialization fails
	ErrTelegramInit = fmt.Errorf("failed to initialize telegram bot api")

	// ErrBotStarted is returned when trying to modify configuration of a running bot
	ErrBotStarted = fmt.Errorf("cannot modify running bot")

	// ErrStateNotFound is returned when user state is not found in cache
	ErrStateNotFound = fmt.Errorf("user state not found")

	// ErrInvalidStateType is returned when state type assertion fails
	ErrInvalidStateType = fmt.Errorf("invalid state type in cache")

	// ErrStateHandlerNotFound is returned when handler for state is not found
	ErrStateHandlerNotFound = fmt.Errorf("state handler not found")

	// ErrSendMessageFailed is returned when all attempts to send message failed
	ErrSendMessageFailed = fmt.Errorf("all attempts to send message failed")

	// ErrEditMessageFailed is returned when all attempts to edit message failed
	ErrEditMessageFailed = fmt.Errorf("all attempts to edit message failed")

	// ErrDeleteMessageFailed is returned when all attempts to delete message failed
	ErrDeleteMessageFailed = fmt.Errorf("all attempts to delete message failed")

	// ErrEmptyTriggers is returned when messageTriggers and callBackTriggers are empty
	ErrEmptyTriggers = fmt.Errorf("messageTriggers and callBackTriggers are empty")
)

// SFMError represents an error with additional context information
type SFMError struct {
	Err   error
	Value interface{}
}

func (e *SFMError) Error() string {
	return fmt.Sprintf("%v: %v", e.Err, e.Value)
}

func (e *SFMError) Unwrap() error {
	return e.Err
}

func IsSFMError(err error) bool {
	var sfmErr *SFMError
	return errors.As(err, &sfmErr)
}

// NewSFMError creates a new SFM error with context
func NewSFMError(err error, value interface{}) error {
	return &SFMError{
		Err:   err,
		Value: value,
	}
}
