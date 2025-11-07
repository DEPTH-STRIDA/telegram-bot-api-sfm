package events

import (
	"errors"
	"fmt"
	"strings"
	"tgfsm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

// Значения по умолчанию для конфигурации события ввода данных
const (
	PromptText             = "Enter data:"
	SuccessText            = "Data successfully saved!"
	ErrorText              = "Error. Please try again."
	SubmitText             = "Confirm"
	ConfirmInputText       = "You entered:\n\n%s\n\nPress the button below to confirm. In case of an error, enter the data again."
	DataNotFoundText       = "Data not found. Please enter the data again."
	DataRetrievalErrorText = "Error retrieving data. Please enter the data again."
	InvalidDataText        = "Invalid data: %s\nPlease enter the data again."
	ProcessingErrorText    = "Error processing data: %s"
	ValidationErrorText    = "Invalid data: %s\nPlease try again."
)

var (
	ErrValidatorRequired  = errors.New("validator is required")
	ErrSubmitTextRequired = errors.New("submit text is required")
)

// ===== Опции для EnterDataConfig =====

// EnterDataOption опция для конфигурации события ввода данных
type EnterDataOption func(*EnterDataConfig)

// WithSubmitText устанавливает текст кнопки подтверждения ввода данных
// Если не установить, используется значение по умолчанию:
//
//	`SubmitText = "Confirm"`
func WithSubmitText(text string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.SubmitText = text
	}
}

// WithPromptText устанавливает текст запроса ввода данных
// Если не установить, используется значение по умолчанию:
//
//	`PromptText = "Enter data:"`
func WithPromptText(text string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.PromptText = text
	}
}

// WithSuccessText устанавливает текст сообщения об успехе
//
// Если установить пустое значение, то сообщение об успехе не будет отправлено
// Можно использовать для кастомной обработки успешного ввода данных.
//
// Если не установить, используется значение по умолчанию:
//
//	`SuccessText = "Data successfully saved!"`
func WithSuccessText(text string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.SuccessText = text
	}
}

// WithErrorText устанавливает текст сообщения об ошибке.
//
// Если установить пустое значение, то сообщение об ошибке не будет отправлено.
// Можно использовать для кастомной обработки ошибок ввода данных.
//
// Если не установить, используется значение по умолчанию:
//
//	`ErrorText = "Error. Please try again."`
func WithErrorText(text string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.ErrorText = text
	}
}

// WithMessageTriggers устанавливает глобальные триггеры для входа в режим ввода данных.
// При указании пользователь сможет войти в состояние ввода данных, отправив триггер.
//
// Если оставить пустым, то вход в режим ввода данных необходимо реализовать самостоятельно в своих состояниях.
// Посмотрите описание функции NewEnterDataEvent для деталей.
func WithMessageTriggers(triggers ...string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.MessageTriggers = triggers
	}
}

// WithCallbackTriggers устанавливает глобальные триггеры для входа в режим ввода данных.
// При указании пользователь сможет войти в состояние ввода данных, отправив callback.
//
// Если оставить пустым, то вход в режим ввода данных необходимо реализовать самостоятельно в своих состояниях.
// Посмотрите описание функции NewEnterDataEvent для деталей.
func WithCallbackTriggers(triggers ...string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.CallbackTriggers = triggers
	}
}

// WithValidator устанавливает функцию валидации введенных данных.
//
// Если не установить, то ограничения на введенные данные не будут проверяться.
func WithValidator(validator func(value string) error) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.Validator = validator
	}
}

// Выполнит, указанную функцию, после успешной валидации данных.
func WithOnSuccessEnterAction(action func(message string, userId int64) error) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.OnSuccessEnterAction = action
	}
}

// WithConfirmInputText устанавливает текст сообщения после ввода данных.
// Использует форматирование: %s будет заменен на введенный текст.
// Если не установить, используется значение по умолчанию:
//
//	`ConfirmInputText = "You entered:\n\n%s\n\nPress the button below to confirm. In case of an error, enter the data again."`
//
// Если установить пустое значение, то сообщение не будет отправлено.
func WithConfirmInputText(text string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.ConfirmInputText = text
	}
}

// WithDataNotFoundText устанавливает текст сообщения, когда данные не найдены в кеше.
// Если не установить, используется значение по умолчанию:
//
//	`DataNotFoundText = "Data not found. Please enter the data again."`
func WithDataNotFoundText(text string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.DataNotFoundText = text
	}
}

// WithDataRetrievalErrorText устанавливает текст сообщения об ошибке получения данных из кеша.
// Если не установить, используется значение по умолчанию:
//
//	`DataRetrievalErrorText = "Error retrieving data. Please enter the data again."`
func WithDataRetrievalErrorText(text string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.DataRetrievalErrorText = text
	}
}

// WithInvalidDataText устанавливает текст сообщения о некорректных данных.
// Использует форматирование: %s будет заменен на текст ошибки валидации.
// Если не установить, используется значение по умолчанию:
//
//	`InvalidDataText = "Invalid data: %s\nPlease enter the data again."`
func WithInvalidDataText(text string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.InvalidDataText = text
	}
}

// WithProcessingErrorText устанавливает текст сообщения об ошибке обработки данных.
// Использует форматирование: %s будет заменен на текст ошибки.
// Если не установить, используется значение по умолчанию:
//
//	`ProcessingErrorText = "Error processing data: %s"`
func WithProcessingErrorText(text string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.ProcessingErrorText = text
	}
}

// WithValidationErrorText устанавливает формат текста сообщения об ошибке валидации.
// Использует форматирование: %s будет заменен на текст ошибки валидатора.
// Если не установить, используется значение по умолчанию:
//
//	`ValidationErrorText = "%s"` (отправляется только текст ошибки валидатора)
//
// Если установить пустое значение, то сообщение об ошибке валидации не будет отправлено.
func WithValidationErrorText(text string) EnterDataOption {
	return func(config *EnterDataConfig) {
		config.ValidationErrorText = text
	}
}

// EnterDataConfig конфигурация для ввода данных
type EnterDataConfig struct {
	MessageTriggers        []string
	CallbackTriggers       []string
	Validator              func(value string) error
	SubmitText             string
	OnSuccessEnterAction   func(message string, userId int64) error
	PromptText             string
	SuccessText            string
	ErrorText              string
	ConfirmInputText       string
	DataNotFoundText       string
	DataRetrievalErrorText string
	InvalidDataText        string
	ProcessingErrorText    string
	ValidationErrorText    string
}

// NewEnterDataEvent создает цепочку состояний для ввода данных с использованием опций
func NewEnterDataEvent(opts ...EnterDataOption) (map[string]tgfsm.State, error) {
	config := &EnterDataConfig{
		PromptText:             PromptText,
		SuccessText:            SuccessText,
		ErrorText:              ErrorText,
		SubmitText:             SubmitText,
		ConfirmInputText:       ConfirmInputText,
		DataNotFoundText:       DataNotFoundText,
		DataRetrievalErrorText: DataRetrievalErrorText,
		InvalidDataText:        InvalidDataText,
		ProcessingErrorText:    ProcessingErrorText,
		ValidationErrorText:    ValidationErrorText,
	}

	// Применяем все опции
	for _, opt := range opts {
		opt(config)
	}

	// Валидация обязательных полей
	if len(config.MessageTriggers) == 0 && len(config.CallbackTriggers) == 0 {
		return nil, tgfsm.ErrEmptyTriggers
	}

	if config.Validator == nil {
		return nil, ErrValidatorRequired
	}

	if config.SubmitText == "" {
		return nil, ErrSubmitTextRequired
	}

	return buildStates(config)
}

// buildStates создает состояния
func buildStates(config *EnterDataConfig) (map[string]tgfsm.State, error) {
	var enterPhaseStateID = uuid.New().String()
	var cacheKey = uuid.New().String()

	// Фаза ввода данных
	var enterPhase tgfsm.State = tgfsm.State{
		Global: false,
		AtEntranceFunc: &tgfsm.Handler{Handle: func(b *tgfsm.Bot, u tgbotapi.Update) error {
			msg := tgbotapi.NewMessage(u.SentFrom().ID, config.PromptText)
			// Если ConfirmInputText не задан, отправляем клавиатуру сразу с промптом
			if config.ConfirmInputText == "" {
				keyboard := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(config.SubmitText),
					),
				)
				keyboard.OneTimeKeyboard = true
				keyboard.ResizeKeyboard = true
				msg.ReplyMarkup = keyboard
			}
			_, err := b.SendMessage(msg)
			return err
		}},
		CatchAllFunc: &tgfsm.Handler{Handle: func(b *tgfsm.Bot, u tgbotapi.Update) error {
			// Если текст есть, то проверяем его корректность.
			if u.Message != nil && u.Message.Text != "" {
				// Пропускаем текст кнопки подтверждения, чтобы он обработался в MessageHandlers
				// Сравниваем без учета регистра, так как в bot.go текст приводится к нижнему регистру
				if strings.EqualFold(strings.TrimSpace(u.Message.Text), strings.TrimSpace(config.SubmitText)) {
					return nil
				}

				if err := config.Validator(u.Message.Text); err != nil {
					// Отправляем сообщение с ошибкой валидации
					if config.ValidationErrorText != "" {
						msgText := fmt.Sprintf(config.ValidationErrorText, err.Error())
						msg := tgbotapi.NewMessage(u.SentFrom().ID, msgText)
						_, sendErr := b.SendMessage(msg)
						if sendErr != nil {
							return sendErr
						}
					}
					return nil
				}

				b.GetCache().Set(cacheKey, u.Message.Text, b.GetExpiration())

				// Если ConfirmInputText не задан, не отправляем сообщение
				if config.ConfirmInputText == "" {
					return nil
				}

				// Создаем клавиатуру с кнопкой подтверждения
				keyboard := tgbotapi.NewReplyKeyboard(
					tgbotapi.NewKeyboardButtonRow(
						tgbotapi.NewKeyboardButton(config.SubmitText),
					),
				)
				keyboard.OneTimeKeyboard = true
				keyboard.ResizeKeyboard = true

				msgText := fmt.Sprintf(config.ConfirmInputText, u.Message.Text)
				msg := tgbotapi.NewMessage(u.SentFrom().ID, msgText)
				msg.ReplyMarkup = keyboard
				_, err := b.SendMessage(msg)
				if err != nil {
					return err
				}
			}
			return nil
		}},
		MessageHandlers: map[string]tgfsm.Handler{
			strings.ToLower(strings.TrimSpace(config.SubmitText)): {
				Handle: func(b *tgfsm.Bot, u tgbotapi.Update) error {
					// Проверяем наличие данных в кеше
					cachedData, found := b.GetCache().Get(cacheKey)
					if !found {
						msg := tgbotapi.NewMessage(u.SentFrom().ID, config.DataNotFoundText)
						_, err := b.SendMessage(msg)
						return err
					}

					// Получаем данные из кеша
					value, ok := cachedData.(string)
					if !ok {
						msg := tgbotapi.NewMessage(u.SentFrom().ID, config.DataRetrievalErrorText)
						_, err := b.SendMessage(msg)
						return err
					}

					// Проверяем корректность данных еще раз
					if err := config.Validator(value); err != nil {
						msgText := fmt.Sprintf(config.InvalidDataText, err.Error())
						msg := tgbotapi.NewMessage(u.SentFrom().ID, msgText)
						_, sendErr := b.SendMessage(msg)
						if sendErr != nil {
							return sendErr
						}
						return err
					}

					// Выполняем действие с данными
					if err := config.OnSuccessEnterAction(value, u.SentFrom().ID); err != nil {
						msgText := fmt.Sprintf(config.ProcessingErrorText, err.Error())
						msg := tgbotapi.NewMessage(u.SentFrom().ID, msgText)
						_, sendErr := b.SendMessage(msg)
						if sendErr != nil {
							return sendErr
						}
						return err
					}

					// Удаляем данные из кеша после успешной обработки
					b.GetCache().Delete(cacheKey)

					// Отправляем сообщение об успехе
					if config.SuccessText != "" {
						msg := tgbotapi.NewMessage(u.SentFrom().ID, config.SuccessText)
						_, sendErr := b.SendMessage(msg)
						if sendErr != nil {
							return sendErr
						}
					}

					// Возвращаем пользователя в начальное состояние
					return b.SetUserState(u.SentFrom().ID, "")
				},
			},
		},
	}

	// Состояние для перехода в состояние ввода данных
	var enterInState = tgfsm.State{
		Global: true,
	}
	var enterInStateID = uuid.New().String()

	if len(config.MessageTriggers) > 0 {
		enterInState.MessageHandlers = make(map[string]tgfsm.Handler)
		for _, t := range config.MessageTriggers {
			enterInState.MessageHandlers[t] = tgfsm.Handler{Handle: tgfsm.NewSetUserStateImmediateHandler(enterPhaseStateID)}
		}
	}

	if len(config.CallbackTriggers) > 0 {
		enterInState.CallbackHandlers = make(map[string]tgfsm.Handler)
		for _, t := range config.CallbackTriggers {
			enterInState.CallbackHandlers[t] = tgfsm.Handler{Handle: tgfsm.NewSetUserStateImmediateHandler(enterPhaseStateID)}
		}
	}

	return map[string]tgfsm.State{
		enterPhaseStateID: enterPhase,
		enterInStateID:    enterInState,
	}, nil
}
