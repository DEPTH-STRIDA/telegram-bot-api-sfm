package tgbotapisfm

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gocache "github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
)

// Config структура для конфигурации бота
type Config struct {
	Token           string           // Токен бота
	Expiration      time.Duration    // Время хранения состояний пользователя
	CleanupInterval time.Duration    // Интервал очистки кеша
	States          map[string]State // Карта состояний
}

// Bot структура для бота
type Bot struct {
	BotAPI        *tgbotapi.BotAPI // API бота. Экспортируется для доступа к нему из вне
	expiration    time.Duration    // Время хранения состояний пользователя
	limiter       *Limiter         // Лимитер для ограничения количества запросов к API
	cache         *gocache.Cache   // Кеш для хранения состояний пользователей
	logger        *zerolog.Logger  // Логгер для записи событий
	states        map[string]State // Состояния пользователя
	globalStates  []*State         // Состояния, в которые может перейти пользователь из любого другоо
	updateHandler HandlerFunc      // Обработчик, который будет вызываться при получении любого обновления
	mu            sync.RWMutex     // Мьютекс для проверки состояния бота
}

// NewBot конструктор нового бота
func NewBot(config Config) (*Bot, error) {
	// Если карта состояний пуста, то нужно ее иницилизировать, чтобы избежать ошибок
	if config.States == nil {
		config.States = make(map[string]State)
	}
	if config.Expiration < 0 {
		return nil, NewValidationError(ErrNegativeExpiration, config.Expiration)
	}
	if config.CleanupInterval < 0 {
		return nil, NewValidationError(ErrNegativeCleanup, config.CleanupInterval)
	}
	if config.Token == "" {
		return nil, ErrInvalidToken
	}

	// Создаем логгер по умолчанию
	l := zerolog.New(os.Stdout).With().Timestamp().Logger()

	botAPI, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, NewValidationError(ErrTelegramInit, err)
	}

	// Дополнительная map'a глобальных состояний
	globalStates := make([]*State, 0)
	for _, state := range config.States {
		if state.Global {
			globalStates = append(globalStates, &state)
		}
	}

	app := Bot{
		BotAPI:       botAPI,
		limiter:      NewLimiter(),
		cache:        gocache.New(config.Expiration, config.CleanupInterval),
		states:       config.States,
		globalStates: globalStates,
		expiration:   config.Expiration,
		logger:       &l,
	}

	return &app, nil
}

// SetLogger заменяет текущий логгер
// Должен вызываться до Start()
func (b *Bot) SetLogger(logger *zerolog.Logger) error {
	if !b.mu.TryRLock() {
		return NewValidationError(ErrBotStarted, "logger")
	}
	defer b.mu.RUnlock()

	b.logger = logger
	return nil
}

// SetUpdateHandler устанавливает обработчик обновлений
// Должен вызываться до Start()
func (b *Bot) SetUpdateHandler(handler HandlerFunc) error {
	if !b.mu.TryRLock() {
		return NewValidationError(ErrBotStarted, "update handler")
	}
	defer b.mu.RUnlock()

	b.updateHandler = handler
	return nil
}

// Start запускает обработку обновлений в горутине
func (b *Bot) Start(offset, timeout int) {
	if !b.mu.TryLock() {
		b.logger.Warn().Msg("Бот уже запущен")
		return
	}
	b.logger.Info().Msg("Запуск бота")
	go b.HandleUpdates(offset, timeout)
}

// Stop останавливает обработку обновлений
func (b *Bot) Stop() {
	b.mu.Unlock()                   // Разблокируем мьютекс, заблокированный в Start()
	b.BotAPI.StopReceivingUpdates() // Останавливаем получение обновлений
	b.logger.Info().Msg("Остановка обработки обновлений")
}

// HandleUpdates запускает обработку всех обновлений поступающих боту из телеграмма
func (app *Bot) HandleUpdates(offset, timeout int) {
	// Настройка обновлений
	u := tgbotapi.NewUpdate(offset)
	u.Timeout = timeout
	updates := app.BotAPI.GetUpdatesChan(u)
	app.logger.Info().Msg("Запуск обработки обновлений")
	for update := range updates {

		// Обработка любого обновления
		go func() {
			if app.updateHandler != nil {
				app.updateHandler(app, update)
			}
		}()

		go func(update tgbotapi.Update) {

			// Обработка локальных стейтов
			if update.SentFrom() == nil {
				return
			}

			// Обработка глобальных стейтов
			globalStateFound, err := app.HandleGlobalStates(update)
			if err != nil {
				app.logger.Error().Err(err).Msg("failed to handle global state")
			}
			// Если глобальное состояние найдено, то выходим из функции
			if globalStateFound {
				return
			}
			// Получение названия состояния пользователя
			userStateName, err := app.GetUserState(update.SentFrom().ID)
			if err != nil {
				return
			}
			// Получени состояния
			userState, ok := app.states[userStateName]
			if !ok {
				app.logger.Error().Str("state", userStateName).Msg("state not found in states map")
			}
			// Обработка обновления по локальному состоянию
			_, err = app.SelectHandler(update, &userState)
			if err != nil {
				app.logger.Error().Err(err).Msg("failed to handle user state")
			}

		}(update)

	}
}

// GetUserState возвращает название состояния, в котором находится пользователь
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

// SetUserState меняет состояние пользователя
func (app *Bot) SetUserState(userId int64, state string) error {
	_, ok := app.states[state]
	if !ok {
		return NewValidationError(ErrStateHandlerNotFound, state)
	}

	app.cache.Set(strconv.FormatInt(userId, 10), state, app.expiration)
	return nil
}

// SetUserStateImmediate меняет состояние пользователя и сразу обрабатывает текущее обновление
func (app *Bot) SetUserStateImmediate(userId int64, state string, update tgbotapi.Update) error {
	if err := app.SetUserState(userId, state); err != nil {
		return err
	}

	if newState, ok := app.states[state]; ok {
		// Вызываем действие при входе, если оно есть и это не глобальное состояние
		if newState.AtEntranceFunc != nil {
			if err := newState.AtEntranceFunc.Handle(app, update); err != nil {
				app.logger.Error().
					Err(err).
					Str("state", state).
					Msg("failed to handle entrance function")
			}
		}

		// Немедленная обработка текущего обновления
		_, err := app.SelectHandler(update, &newState)
		if err != nil {
			app.logger.Error().
				Err(err).
				Str("state", state).
				Msg("failed to handle immediate reaction")
		}
	}

	return nil
}

// HandleGlobalStates проверяет подходит ли действие пользователя под
// глобальные состояния и если подходит, то выполняет его.
// Возвращает true, если обработчик нашелся и выполнился.
func (app *Bot) HandleGlobalStates(update tgbotapi.Update) (bool, error) {
	// Обработка всех глобальных состояний
	for _, state := range app.globalStates {
		// Обработка состояния
		handlerIsFound, err := app.SelectHandler(update, state)
		// Если ошибка, то пропускаем состояние
		if err != nil {
			app.logger.Error().Err(err).Msg("failed to handle global state")
			continue
		}
		// Если обработчик найден, то возвращаем true
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
			app.logger.Info().
				Str("command", update.Message.Text).
				Int64("chat_id", update.Message.Chat.ID).
				Str("username", update.Message.Chat.UserName).
				Msg("command not found")
			return false, nil
		}
	case update.CallbackQuery != nil:
		if userState.CallbackHandlers != nil {
			return app.handleCallback(userState, update)
		} else {
			app.logger.Info().
				Str("callback", update.CallbackQuery.Data).
				Int64("user_id", update.CallbackQuery.From.ID).
				Str("username", update.CallbackQuery.From.UserName).
				Msg("callback not found")
			return false, nil
		}
	}
	return false, nil
}

// handleMessage ищет команду в map'е и выполняет ее
func (app *Bot) handleMessage(userState *State, update tgbotapi.Update) (bool, error) {
	messageFound := false

	// Поиск обработчика
	if currentAction, ok := userState.MessageHandlers[strings.ToLower(strings.TrimSpace(update.Message.Text))]; ok {
		messageFound = true
		if err := currentAction.Handle(app, update); err != nil {
			app.logger.Error().
				Err(err).
				Str("command", update.Message.Text).
				Int64("chat_id", update.Message.Chat.ID).
				Str("username", update.Message.Chat.UserName).
				Msg("failed to handle command")
		} else {
			app.logger.Info().
				Str("command", update.Message.Text).
				Int64("chat_id", update.Message.Chat.ID).
				Str("username", update.Message.Chat.UserName).
				Msg("command handled successfully")
		}
	} else {
		if userState.CatchAllFunc != nil {
			err := userState.CatchAllFunc.Handle(app, update)
			if err != nil {
				app.logger.Error().
					Err(err).
					Int64("chat_id", update.Message.Chat.ID).
					Str("username", update.Message.Chat.UserName).
					Str("command", update.Message.Text).
					Msg("failed to handle command")
			}
		} else {
			app.logger.Info().
				Int64("chat_id", update.Message.Chat.ID).
				Str("username", update.Message.Chat.UserName).
				Str("command", update.Message.Text).
				Msg("command not found")
		}

	}
	return messageFound, nil
}

// handleCallback ищет команду в map'е и выполняет ее
func (app *Bot) handleCallback(userState *State, update tgbotapi.Update) (bool, error) {
	callbackFound := false

	if currentAction, ok := userState.CallbackHandlers[update.CallbackQuery.Data]; ok {
		callbackFound = true
		if err := currentAction.Handle(app, update); err != nil {
			app.logger.Error().
				Err(err).
				Str("callback", update.CallbackQuery.Data).
				Int64("user_id", update.CallbackQuery.From.ID).
				Str("username", update.CallbackQuery.From.UserName).
				Msg("failed to handle callback")
			return callbackFound, err
		}

		app.logger.Info().
			Str("callback", update.CallbackQuery.Data).
			Int64("user_id", update.CallbackQuery.From.ID).
			Str("username", update.CallbackQuery.From.UserName).
			Msg("callback handled successfully")
	} else {
		if userState.CatchAllFunc != nil {
			err := userState.CatchAllFunc.Handle(app, update)
			if err != nil {
				app.logger.Error().
					Err(err).
					Int64("user_id", update.CallbackQuery.From.ID).
					Str("username", update.CallbackQuery.From.UserName).
					Str("callback", update.CallbackQuery.Data).
					Msg("failed to handle callback")
				return callbackFound, err
			}
		} else {
			app.logger.Info().
				Int64("user_id", update.CallbackQuery.From.ID).
				Str("username", update.CallbackQuery.From.UserName).
				Str("callback", update.CallbackQuery.Data).
				Msg("callback not found")
		}
	}
	return callbackFound, nil
}
