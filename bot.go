package tgfsm

import (
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gocache "github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

const (
	// Default values
	DefaultExpiration      = 24 * time.Hour // Default user state storage duration
	DefaultCleanupInterval = 1 * time.Hour  // Default cache cleanup interval
)

// Bot represents the bot instance
type Bot struct {
	BotAPI            *tgbotapi.BotAPI // Bot API. Exported for external access
	expiration        time.Duration    // User state storage duration
	cleanupInterval   time.Duration    // Cache cleanup interval
	limiter           *Limiter         // Limiter for API request rate limiting
	cache             *gocache.Cache   // Cache for storing user states
	logger            *zap.Logger      // Logger for recording events
	states            map[string]State // User states
	globalStates      []*State         // States that user can transition to from any other state
	updateHandler     HandlerFunc      // Handler that will be called when receiving any update
	privateOnly       bool             // Only accept messages from private chats
	blacklistedChats  map[int64]bool   // Blacklisted chat IDs
	mu                sync.RWMutex     // Mutex for bot state (start/stop/update)
	blacklistMu       sync.RWMutex     // Mutex for blacklist operations
	autoDeleteEnabled bool             // Enable auto-deletion of last message
	lastMessageCache  LastMessageCache // Cache for last message IDs
}

// NewBot creates a new bot instance
func NewBot(token string, options ...Option) (*Bot, error) {
	// Validate token first
	if token == "" {
		return nil, ErrInvalidToken
	}

	// Initialize bot with default values
	app := Bot{
		limiter: NewLimiter(),
	}

	// Apply options
	for _, option := range options {
		option(&app)
	}

	// Set default values if not provided by options
	if err := app.setDefaults(); err != nil {
		return nil, err
	}

	// Validate configured values
	if app.expiration < 0 {
		return nil, NewSFMError(ErrNegativeExpiration, app.expiration)
	}
	if app.cleanupInterval < 0 {
		return nil, NewSFMError(ErrNegativeCleanup, app.cleanupInterval)
	}

	// Initialize Telegram Bot API
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, NewSFMError(ErrTelegramInit, err)
	}
	app.BotAPI = botAPI

	// Initialize cache with configured values
	app.cache = gocache.New(app.expiration, app.cleanupInterval)

	// Build global states map
	globalStates := make([]*State, 0)
	for _, state := range app.states {
		if state.Global {
			globalStates = append(globalStates, &state)
		}
	}
	app.globalStates = globalStates

	return &app, nil
}

// setDefaults sets default values for unconfigured fields
func (b *Bot) setDefaults() error {
	if b.expiration == 0 {
		b.expiration = DefaultExpiration
	}
	if b.cleanupInterval == 0 {
		b.cleanupInterval = DefaultCleanupInterval
	}
	if b.states == nil {
		b.states = make(map[string]State)
	}
	if b.blacklistedChats == nil {
		b.blacklistedChats = make(map[int64]bool)
	}
	if b.logger == nil {
		logger, err := NewZapLogger()
		if err != nil {
			return NewSFMError(ErrTelegramInit, err)
		}
		b.logger = logger
	}

	return nil
}

// UpdateBot updates bot configuration with new options
// Can only be called when bot is not running
func (b *Bot) UpdateBot(options ...Option) error {
	// Try to acquire write lock - this will fail if bot is running
	if !b.mu.TryLock() {
		return NewSFMError(ErrBotStarted, "cannot update running bot")
	}
	defer b.mu.Unlock()

	// Apply new options
	for _, option := range options {
		option(b)
	}

	// Set defaults for any unset values
	if err := b.setDefaults(); err != nil {
		return err
	}

	// Validate configured values
	if b.expiration < 0 {
		return NewSFMError(ErrNegativeExpiration, b.expiration)
	}
	if b.cleanupInterval < 0 {
		return NewSFMError(ErrNegativeCleanup, b.cleanupInterval)
	}

	// Reinitialize cache with new values
	b.cache = gocache.New(b.expiration, b.cleanupInterval)

	// Rebuild global states map
	globalStates := make([]*State, 0)
	for _, state := range b.states {
		if state.Global {
			globalStates = append(globalStates, &state)
		}
	}
	b.globalStates = globalStates

	return nil
}

// AddToBlacklist adds a chat ID to the blacklist
// Can be called while bot is running (uses blacklistMu)
func (b *Bot) AddToBlacklist(chatID int64) {
	b.blacklistMu.Lock()
	defer b.blacklistMu.Unlock()
	b.blacklistedChats[chatID] = true
}

// RemoveFromBlacklist removes a chat ID from the blacklist
// Can be called while bot is running (uses blacklistMu)
func (b *Bot) RemoveFromBlacklist(chatID int64) {
	b.blacklistMu.Lock()
	defer b.blacklistMu.Unlock()
	delete(b.blacklistedChats, chatID)
}

// IsBlacklisted checks if a chat ID is blacklisted
// Can be called while bot is running (uses blacklistMu)
func (b *Bot) IsBlacklisted(chatID int64) bool {
	b.blacklistMu.RLock()
	defer b.blacklistMu.RUnlock()
	return b.blacklistedChats[chatID]
}

// GetBlacklist returns a copy of the current blacklist
// Can be called while bot is running (uses blacklistMu)
func (b *Bot) GetBlacklist() []int64 {
	b.blacklistMu.RLock()
	defer b.blacklistMu.RUnlock()

	blacklist := make([]int64, 0, len(b.blacklistedChats))
	for chatID := range b.blacklistedChats {
		blacklist = append(blacklist, chatID)
	}
	return blacklist
}

// IsPrivateOnly returns whether bot only accepts private messages
// Can be called while bot is running (uses main mu)
func (b *Bot) IsPrivateOnly() bool {
	return b.privateOnly
}

// shouldProcessUpdate determines if an update should be processed based on filters
func (b *Bot) shouldProcessUpdate(update tgbotapi.Update) bool {
	// Check if update has a chat
	var chatID int64
	if update.Message != nil {
		chatID = update.Message.Chat.ID
	} else if update.CallbackQuery != nil {
		chatID = update.CallbackQuery.Message.Chat.ID
	} else {
		// No chat information, skip
		return false
	}

	// Check if chat is blacklisted (uses blacklistMu)
	if b.IsBlacklisted(chatID) {
		return false
	}
	// Check if private only mode is enabled (uses main mu)
	if b.IsPrivateOnly() {
		// Only process if it's a private chat
		if update.Message != nil {
			return update.Message.Chat.Type == "private"
		} else if update.CallbackQuery != nil {
			return update.CallbackQuery.Message.Chat.Type == "private"
		}
	}

	return true
}

// Start starts update processing in a goroutine
func (b *Bot) Start(offset, timeout int) {
	if !b.mu.TryLock() {
		b.logger.Warn("Bot is already running")
		return
	}
	b.logger.Info("Starting bot")
	go b.HandleUpdates(offset, timeout)
	// Note: mutex remains locked while bot is running
}

// Stop stops update processing
func (b *Bot) Stop() {
	b.BotAPI.StopReceivingUpdates() // Stop receiving updates
	b.mu.Unlock()                   // Unlock mutex locked in Start()
	b.logger.Info("Stopping update processing")
}

// HandleUpdates starts processing all updates received by the bot from Telegram
func (app *Bot) HandleUpdates(offset, timeout int) {
	// Configure updates
	u := tgbotapi.NewUpdate(offset)
	u.Timeout = timeout
	updates := app.BotAPI.GetUpdatesChan(u)
	app.logger.Info("Starting update processing")

	for update := range updates {
		// Check if update should be processed based on filters
		if !app.shouldProcessUpdate(update) {
			continue
		}
		go func(update tgbotapi.Update) {
			if app.updateHandler != nil {
				app.updateHandler(app, update)
			}

			// Process local states
			if update.SentFrom() == nil {
				return
			}

			// Process global states
			globalStateFound, err := app.HandleGlobalStates(update)
			if err != nil {
				app.logger.Error("failed to handle global state", zap.Error(err))
			}
			// If global state is found, exit the function
			if globalStateFound {
				return
			}
			// Get user state name
			userStateName, err := app.GetUserState(update.SentFrom().ID)
			if err != nil {
				return
			}
			// Get state
			userState, ok := app.states[userStateName]
			if !ok {
				app.logger.Error("state not found in states map", zap.String("state", userStateName))
			}
			// Process update by local state
			_, err = app.SelectHandler(update, &userState)
			if err != nil {
				app.logger.Error("failed to handle user state", zap.Error(err))
			}

		}(update)

	}
}

// GetUserState returns the name of the state the user is currently in
func (app *Bot) GetUserState(userId int64) (string, error) {
	userStateInterface, ok := app.cache.Get(strconv.FormatInt(userId, 10))
	if !ok {
		return "", ErrStateNotFound
	}

	userState, ok := userStateInterface.(string)
	if !ok {
		return "", ErrInvalidStateType
	}

	return userState, nil
}

// SetUserState changes the user's state
func (app *Bot) SetUserState(userId int64, state string) error {
	_, ok := app.states[state]
	if !ok {
		return NewSFMError(ErrStateHandlerNotFound, state)
	}

	app.cache.Set(strconv.FormatInt(userId, 10), state, app.expiration)
	return nil
}

// SetUserStateImmediate changes the user's state and immediately processes the current update
func (app *Bot) SetUserStateImmediate(userId int64, state string, update tgbotapi.Update) error {
	if err := app.SetUserState(userId, state); err != nil {
		return err
	}

	if newState, ok := app.states[state]; ok {
		// Call entrance action if it exists and this is not a global state
		if newState.AtEntranceFunc != nil {
			if err := newState.AtEntranceFunc.Handle(app, update); err != nil {
				app.logger.Error("failed to handle entrance function", zap.Error(err))
			}
			return nil
		}

		// Immediate processing of current update
		_, err := app.SelectHandler(update, &newState)
		if err != nil {
			app.logger.Error("failed to handle immediate reaction", zap.Error(err))
		}
	}

	return nil
}

// HandleGlobalStates checks if user action matches global states and executes it if it does.
// Returns true if a handler was found and executed.
func (app *Bot) HandleGlobalStates(update tgbotapi.Update) (bool, error) {
	// Process all global states
	for _, state := range app.globalStates {
		// Process state

		handlerIsFound, err := app.SelectHandler(update, state)
		// If error, skip state
		if err != nil {
			app.logger.Error("failed to handle global state", zap.Error(err))
			continue
		}
		// If handler found, return true
		if handlerIsFound {
			return true, nil
		}
	}
	return false, nil
}

func (app *Bot) SelectHandler(update tgbotapi.Update, userState *State) (bool, error) {
	switch {
	case update.Message != nil:
		if userState.MessageHandlers != nil {
			return app.handleMessage(userState, update)
		} else {
			app.logger.Info("command not found",
				zap.String("command", update.Message.Text),
				zap.Int64("chat_id", update.Message.Chat.ID),
				zap.String("username", update.Message.Chat.UserName),
			)
			return false, nil
		}
	case update.CallbackQuery != nil:
		if userState.CallbackHandlers != nil {
			return app.handleCallback(userState, update)
		} else {
			app.logger.Info("callback not found",
				zap.String("callback", update.CallbackQuery.Data),
				zap.Int64("user_id", update.CallbackQuery.From.ID),
				zap.String("username", update.CallbackQuery.From.UserName),
			)
			return false, nil
		}
	}
	return false, nil
}

// handleMessage searches for a command in the map and executes it
func (app *Bot) handleMessage(userState *State, update tgbotapi.Update) (bool, error) {
	messageFound := false

	// Search for handler
	if currentAction, ok := userState.MessageHandlers[strings.ToLower(strings.TrimSpace(update.Message.Text))]; ok {
		messageFound = true
		if err := currentAction.Handle(app, update); err != nil {
			app.logger.Error("failed to handle command", zap.Error(err))
		} else {
			app.logger.Info("command handled successfully",
				zap.String("command", update.Message.Text),
				zap.Int64("chat_id", update.Message.Chat.ID),
				zap.String("username", update.Message.Chat.UserName),
			)
		}
	} else {
		if userState.CatchAllFunc != nil {
			err := userState.CatchAllFunc.Handle(app, update)
			if err != nil {
				app.logger.Error("failed to handle command", zap.Error(err))
			}
		} else {
			app.logger.Info("command not found",
				zap.String("command", update.Message.Text),
				zap.Int64("chat_id", update.Message.Chat.ID),
				zap.String("username", update.Message.Chat.UserName),
			)
		}

	}
	return messageFound, nil
}

// handleCallback searches for a command in the map and executes it
func (app *Bot) handleCallback(userState *State, update tgbotapi.Update) (bool, error) {
	callbackFound := false

	if currentAction, ok := userState.CallbackHandlers[update.CallbackQuery.Data]; ok {
		callbackFound = true
		if err := currentAction.Handle(app, update); err != nil {
			app.logger.Error("failed to handle callback", zap.Error(err))
			return callbackFound, err
		}

		app.logger.Info("callback handled successfully",
			zap.String("callback", update.CallbackQuery.Data),
			zap.Int64("user_id", update.CallbackQuery.From.ID),
			zap.String("username", update.CallbackQuery.From.UserName),
		)
	} else {
		if userState.CatchAllFunc != nil {
			err := userState.CatchAllFunc.Handle(app, update)
			if err != nil {
				app.logger.Error("failed to handle callback", zap.Error(err))
			}
		} else {
			app.logger.Info("callback not found",
				zap.Int64("user_id", update.CallbackQuery.From.ID),
				zap.String("username", update.CallbackQuery.From.UserName),
				zap.String("callback", update.CallbackQuery.Data),
			)
		}
	}
	return callbackFound, nil
}

func (app *Bot) GetExpiration() time.Duration {
	return app.expiration
}

func (app *Bot) GetCache() *gocache.Cache {
	return app.cache
}
