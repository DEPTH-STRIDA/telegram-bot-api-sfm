package events

import (
	"errors"
	"tgfsm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

// Значения по умолчанию для конфигурации простого слайдера
const (
	SimpleSliderPrevButtonText = "◀️"
	SimpleSliderNextButtonText = "▶️"
)

var (
	ErrEmptyTexts = errors.New("texts array is required and cannot be empty")
)

// ===== Опции для SimpleSliderConfig =====

// SimpleSliderOption опция для конфигурации простого слайдера
type SimpleSliderOption func(*SimpleSliderConfig)

// SimpleSliderButton представляет дополнительную кнопку для слайдера
type SimpleSliderButton struct {
	Text     string
	Callback string
}

// SimpleSliderConfig конфигурация для простого слайдера
type SimpleSliderConfig struct {
	MessageTriggers   []string
	CallbackTriggers  []string
	Texts             []string
	PrevButtonText    string
	NextButtonText    string
	AdditionalButtons []SimpleSliderButton
	CurrentIndexKey   string
	MessageIDKey      string
}

// WithSimpleSliderMessageTriggers устанавливает глобальные триггеры для входа в режим слайдера.
// При указании пользователь сможет войти в состояние слайдера, отправив триггер.
func WithSimpleSliderMessageTriggers(triggers ...string) SimpleSliderOption {
	return func(config *SimpleSliderConfig) {
		config.MessageTriggers = triggers
	}
}

// WithSimpleSliderCallbackTriggers устанавливает глобальные триггеры для входа в режим слайдера.
// При указании пользователь сможет войти в состояние слайдера, отправив callback.
func WithSimpleSliderCallbackTriggers(triggers ...string) SimpleSliderOption {
	return func(config *SimpleSliderConfig) {
		config.CallbackTriggers = triggers
	}
}

// WithSimpleSliderTexts устанавливает массив текстов для листания.
// Обязательный параметр.
func WithSimpleSliderTexts(texts ...string) SimpleSliderOption {
	return func(config *SimpleSliderConfig) {
		config.Texts = texts
	}
}

// WithSimpleSliderPrevButtonText устанавливает текст кнопки "Назад".
// Если не установить, используется значение по умолчанию:
//
//	`SimpleSliderPrevButtonText = "◀️ Назад"`
func WithSimpleSliderPrevButtonText(text string) SimpleSliderOption {
	return func(config *SimpleSliderConfig) {
		config.PrevButtonText = text
	}
}

// WithSimpleSliderNextButtonText устанавливает текст кнопки "Вперед".
// Если не установить, используется значение по умолчанию:
//
//	`SimpleSliderNextButtonText = "Вперед ▶️"`
func WithSimpleSliderNextButtonText(text string) SimpleSliderOption {
	return func(config *SimpleSliderConfig) {
		config.NextButtonText = text
	}
}

// WithSimpleSliderAdditionalButtons устанавливает дополнительные кнопки для слайдера.
// Кнопки будут отображаться на отдельной строке после кнопок листания.
// Каждая кнопка должна иметь текст и callback данные.
//
// Пример:
//
//	WithSimpleSliderAdditionalButtons(
//		SimpleSliderButton{Text: "Главное меню", Callback: "menu"},
//		SimpleSliderButton{Text: "Назад", Callback: "back"},
//	)
func WithSimpleSliderAdditionalButtons(buttons ...SimpleSliderButton) SimpleSliderOption {
	return func(config *SimpleSliderConfig) {
		config.AdditionalButtons = buttons
	}
}

// NewSimpleSliderEvent создает цепочку состояний для простого слайдера с использованием опций
func NewSimpleSliderEvent(opts ...SimpleSliderOption) (map[string]tgfsm.State, error) {
	config := &SimpleSliderConfig{
		PrevButtonText:  SimpleSliderPrevButtonText,
		NextButtonText:  SimpleSliderNextButtonText,
		CurrentIndexKey: uuid.New().String(),
		MessageIDKey:    uuid.New().String(),
	}

	// Применяем все опции
	for _, opt := range opts {
		opt(config)
	}

	// Валидация обязательных полей
	if len(config.MessageTriggers) == 0 && len(config.CallbackTriggers) == 0 {
		return nil, tgfsm.ErrEmptyTriggers
	}

	if len(config.Texts) == 0 {
		return nil, ErrEmptyTexts
	}

	return buildSimpleSliderStates(config)
}

// buildSimpleSliderStates создает состояния для простого слайдера
func buildSimpleSliderStates(config *SimpleSliderConfig) (map[string]tgfsm.State, error) {
	var sliderStateID = uuid.New().String()

	// Фаза листания текстов
	var sliderPhase tgfsm.State = tgfsm.State{
		Global: false,
		AtEntranceFunc: &tgfsm.Handler{Handle: func(b *tgfsm.Bot, u tgbotapi.Update) error {
			// Устанавливаем начальный индекс
			b.GetCache().Set(config.CurrentIndexKey, 0, b.GetExpiration())

			// Отправляем первое сообщение
			return sendSliderMessage(b, u, config, 0)
		}},
		CallbackHandlers: map[string]tgfsm.Handler{
			"prev": {
				Handle: func(b *tgfsm.Bot, u tgbotapi.Update) error {
					// Отвечаем на callback query
					if u.CallbackQuery != nil {
						callback := tgbotapi.NewCallback(u.CallbackQuery.ID, "")
						b.BotAPI.Request(callback)
					}

					// Получаем текущий индекс
					cachedIndex, found := b.GetCache().Get(config.CurrentIndexKey)
					if !found {
						// Если индекс не найден, начинаем с начала
						b.GetCache().Set(config.CurrentIndexKey, 0, b.GetExpiration())
						return sendSliderMessage(b, u, config, 0)
					}

					currentIndex, ok := cachedIndex.(int)
					if !ok {
						currentIndex = 0
					}

					// Уменьшаем индекс
					if currentIndex > 0 {
						currentIndex--
					}

					b.GetCache().Set(config.CurrentIndexKey, currentIndex, b.GetExpiration())

					// Обновляем сообщение
					return updateSliderMessage(b, u, config, currentIndex)
				},
			},
			"next": {
				Handle: func(b *tgfsm.Bot, u tgbotapi.Update) error {
					// Отвечаем на callback query
					if u.CallbackQuery != nil {
						callback := tgbotapi.NewCallback(u.CallbackQuery.ID, "")
						b.BotAPI.Request(callback)
					}

					// Получаем текущий индекс
					cachedIndex, found := b.GetCache().Get(config.CurrentIndexKey)
					if !found {
						// Если индекс не найден, начинаем с начала
						b.GetCache().Set(config.CurrentIndexKey, 0, b.GetExpiration())
						return sendSliderMessage(b, u, config, 0)
					}

					currentIndex, ok := cachedIndex.(int)
					if !ok {
						currentIndex = 0
					}

					// Увеличиваем индекс
					if currentIndex < len(config.Texts)-1 {
						currentIndex++
					}

					b.GetCache().Set(config.CurrentIndexKey, currentIndex, b.GetExpiration())

					// Обновляем сообщение
					return updateSliderMessage(b, u, config, currentIndex)
				},
			},
		},
	}

	// Состояние для перехода в состояние слайдера
	var enterInState = tgfsm.State{
		Global: true,
	}
	var enterInStateID = uuid.New().String()

	if len(config.MessageTriggers) > 0 {
		enterInState.MessageHandlers = make(map[string]tgfsm.Handler)
		for _, t := range config.MessageTriggers {
			enterInState.MessageHandlers[t] = tgfsm.Handler{Handle: tgfsm.NewSetUserStateImmediateHandler(sliderStateID)}
		}
	}

	if len(config.CallbackTriggers) > 0 {
		enterInState.CallbackHandlers = make(map[string]tgfsm.Handler)
		for _, t := range config.CallbackTriggers {
			enterInState.CallbackHandlers[t] = tgfsm.Handler{Handle: tgfsm.NewSetUserStateImmediateHandler(sliderStateID)}
		}
	}

	return map[string]tgfsm.State{
		sliderStateID:  sliderPhase,
		enterInStateID: enterInState,
	}, nil
}

// sendSliderMessage отправляет новое сообщение со слайдером
func sendSliderMessage(b *tgfsm.Bot, u tgbotapi.Update, config *SimpleSliderConfig, index int) error {
	text := config.Texts[index]
	keyboard := buildSliderKeyboard(config, index)

	msg := tgbotapi.NewMessage(u.SentFrom().ID, text)
	msg.ReplyMarkup = keyboard
	sentMsg, err := b.SendMessage(msg)
	if err != nil {
		return err
	}

	// Сохраняем ID сообщения для последующего обновления
	b.GetCache().Set(config.MessageIDKey, sentMsg.MessageID, b.GetExpiration())

	return nil
}

// updateSliderMessage обновляет существующее сообщение со слайдером
func updateSliderMessage(b *tgfsm.Bot, u tgbotapi.Update, config *SimpleSliderConfig, index int) error {
	text := config.Texts[index]
	keyboard := buildSliderKeyboard(config, index)

	var msgID int
	var chatID int64

	// Пытаемся получить ID сообщения из callback query
	if u.CallbackQuery != nil && u.CallbackQuery.Message != nil {
		msgID = u.CallbackQuery.Message.MessageID
		chatID = u.CallbackQuery.Message.Chat.ID
	} else {
		// Если нет callback query, получаем из кеша
		cachedMsgID, found := b.GetCache().Get(config.MessageIDKey)
		if !found {
			// Если ID не найден, отправляем новое сообщение
			return sendSliderMessage(b, u, config, index)
		}

		var ok bool
		msgID, ok = cachedMsgID.(int)
		if !ok {
			// Если ID неверного типа, отправляем новое сообщение
			return sendSliderMessage(b, u, config, index)
		}
		chatID = u.SentFrom().ID
	}

	// Обновляем сообщение
	editMsg := tgbotapi.NewEditMessageText(chatID, msgID, text)
	editMsg.ReplyMarkup = keyboard
	_, err := b.EditMessage(editMsg)
	return err
}

// buildSliderKeyboard создает клавиатуру для слайдера
func buildSliderKeyboard(config *SimpleSliderConfig, index int) *tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton

	// Определяем, какие кнопки показывать
	isFirst := index == 0
	isLast := index == len(config.Texts)-1

	// Первая строка - кнопки листания
	var navigationButtons []tgbotapi.InlineKeyboardButton

	// Если не первый элемент, показываем кнопку "Назад"
	if !isFirst {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData(config.PrevButtonText, "prev"))
	}

	// Если не последний элемент, показываем кнопку "Вперед"
	if !isLast {
		navigationButtons = append(navigationButtons, tgbotapi.NewInlineKeyboardButtonData(config.NextButtonText, "next"))
	}

	// Если есть кнопки навигации, добавляем их в первый ряд
	if len(navigationButtons) > 0 {
		rows = append(rows, navigationButtons)
	}

	// Добавляем дополнительные кнопки, если они есть
	if len(config.AdditionalButtons) > 0 {
		var additionalButtons []tgbotapi.InlineKeyboardButton
		for _, btn := range config.AdditionalButtons {
			additionalButtons = append(additionalButtons, tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.Callback))
		}
		rows = append(rows, additionalButtons)
	}

	return &tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: rows,
	}
}
