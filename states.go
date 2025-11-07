package tgfsm

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// HandlerFunc is a function type for message handlers
type HandlerFunc func(b *Bot, u tgbotapi.Update) error

// Handler represents a message handler with its metadata
type Handler struct {
	// Handle processes incoming updates from Telegram
	Handle HandlerFunc

	// Description provides information about the handler
	Description *string
}

// State represents a bot state and defines message processing rules
type State struct {
	// If true, handler triggers are checked independently of the user's current state
	// If a suitable handler is found in a global state, it is executed and triggers of other
	// states are not executed.
	// After the first matching global state, other global states are not executed.
	// Global states are set once during initialization.
	Global bool
	// Executed when entering the state. Every time an update comes from the user while the user is in this state.
	AtEntranceFunc *Handler
	// Executed for all events that did not match any routes.
	CatchAllFunc *Handler
	// Maps message text to handler key and executes it.
	// User text is converted to lowercase, so keys should be in the same format.
	MessageHandlers map[string]Handler
	// Maps callback data to handler key and executes it
	CallbackHandlers map[string]Handler
}

// NewState creates a new State instance with the given parameters
func NewState(global bool, atEntranceFunc *Handler, catchAllFunc *Handler) *State {
	return &State{
		Global:           global,
		AtEntranceFunc:   atEntranceFunc,
		CatchAllFunc:     catchAllFunc,
		MessageHandlers:  make(map[string]Handler),
		CallbackHandlers: make(map[string]Handler),
	}
}

// NewSetUserStateHandler создает обработчик для установки состояния пользователя
func NewSetUserStateHandler(state string) HandlerFunc {
	return func(b *Bot, u tgbotapi.Update) error {
		return b.SetUserState(u.SentFrom().ID, state)
	}
}

// NewSetUserStateImmediateHandler создает обработчик для установки состояния пользователя с мгновенной обработкой этого состояние
// Пример использования:
// state := NewSetUserStateImmediateHandler("state")
// state.Handle(b, u)
// Событие 'u' будет сразу обработано AtEntranceFunc и MessageHandlers/CallbackHandlers, при наличии этих обработчиков.
func NewSetUserStateImmediateHandler(state string) HandlerFunc {
	return func(b *Bot, u tgbotapi.Update) error {
		return b.SetUserStateImmediate(u.SentFrom().ID, state, u)
	}
}
