package events

// // SliderConfig конфигурация для слайдера
// type SliderConfig struct {
// 	MessageTriggers  []string
// 	CallbackTriggers []string
// 	Validator        func(value string) error
// 	SubmitText       string
// 	OnEnterAction    func(message string, userId int64) error
// 	PromptText       string
// 	SuccessText      string
// 	ErrorText        string
// 	MinValue         int
// 	MaxValue         int
// 	Step             int
// }

// // ===== Опции для SliderConfig =====

// type SliderOption func(*SliderConfig)

// func WithSliderSubmitText(text string) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.SubmitText = text
// 	}
// }

// func WithSliderPromptText(text string) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.PromptText = text
// 	}
// }

// func WithSliderSuccessText(text string) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.SuccessText = text
// 	}
// }

// func WithSliderErrorText(text string) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.ErrorText = text
// 	}
// }

// func WithSliderMessageTriggers(triggers ...string) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.MessageTriggers = triggers
// 	}
// }

// func WithSliderCallbackTriggers(triggers ...string) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.CallbackTriggers = triggers
// 	}
// }

// func WithSliderValidator(validator func(value string) error) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.Validator = validator
// 	}
// }

// func WithSliderOnEnterAction(action func(message string, userId int64) error) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.OnEnterAction = action
// 	}
// }

// func WithMinValue(value int) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.MinValue = value
// 	}
// }

// func WithMaxValue(value int) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.MaxValue = value
// 	}
// }

// func WithStep(step int) SliderOption {
// 	return func(config *SliderConfig) {
// 		config.Step = step
// 	}
// }

// // NewSliderEvent создает событие слайдера
// func NewSliderEvent(opts ...SliderOption) (map[string]tgfsm.State, error) {
// 	config := &SliderConfig{
// 		PromptText:  "Выберите значение:",
// 		SuccessText: "Значение успешно установлено!",
// 		ErrorText:   "Произошла ошибка при установке значения.",
// 		MinValue:    0,
// 		MaxValue:    100,
// 		Step:        1,
// 	}

// 	// Применяем все опции
// 	for _, opt := range opts {
// 		opt(config)
// 	}

// 	// Валидация обязательных полей
// 	if len(config.MessageTriggers) == 0 && len(config.CallbackTriggers) == 0 {
// 		return nil, tgfsm.ErrEmptyTriggers
// 	}

// 	if config.Validator == nil {
// 		return nil, fmt.Errorf("validator is required")
// 	}

// 	if config.SubmitText == "" {
// 		return nil, fmt.Errorf("submit text is required")
// 	}

// 	if config.OnEnterAction == nil {
// 		return nil, fmt.Errorf("on enter action is required")
// 	}

// 	return buildSliderStates(config)
// }

// // buildSliderStates создает состояния для слайдера
// func buildSliderStates(config *SliderConfig) (map[string]tgfsm.State, error) {
// 	var sliderStateID = uuid.New().String()
// 	var cacheKey = uuid.New().String()

// 	// Фаза выбора значения
// 	var sliderPhase tgfsm.State = tgfsm.State{
// 		Global: false,
// 		AtEntranceFunc: &tgfsm.Handler{Handle: func(b *tgfsm.Bot, u tgbotapi.Update) error {
// 			// Создаем inline клавиатуру со слайдером
// 			keyboard := tgbotapi.NewInlineKeyboardMarkup(
// 				tgbotapi.NewInlineKeyboardRow(
// 					tgbotapi.NewInlineKeyboardButtonData("◀️", "decrease"),
// 					tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d", config.MinValue), "current"),
// 					tgbotapi.NewInlineKeyboardButtonData("▶️", "increase"),
// 				),
// 				tgbotapi.NewInlineKeyboardRow(
// 					tgbotapi.NewInlineKeyboardButtonData("✅ Подтвердить", config.SubmitText),
// 				),
// 			)

// 			msg := tgbotapi.NewMessage(u.SentFrom().ID, config.PromptText)
// 			msg.ReplyMarkup = keyboard
// 			_, err := b.SendMessage(msg)
// 			return err
// 		}},
// 		CallbackHandlers: map[string]tgfsm.Handler{
// 			"increase": {
// 				Handle: func(b *tgfsm.Bot, u tgbotapi.Update) error {
// 					// Логика увеличения значения
// 					return nil
// 				},
// 			},
// 			"decrease": {
// 				Handle: func(b *tgfsm.Bot, u tgbotapi.Update) error {
// 					// Логика уменьшения значения
// 					return nil
// 				},
// 			},
// 			config.SubmitText: {
// 				Handle: func(b *tgfsm.Bot, u tgbotapi.Update) error {
// 					// Обработка подтверждения
// 					cachedData, found := b.GetCache().Get(cacheKey)
// 					if !found {
// 						msg := tgbotapi.NewMessage(u.SentFrom().ID, "Значение не выбрано. Пожалуйста, выберите значение.")
// 						_, err := b.SendMessage(msg)
// 						return err
// 					}

// 					value, ok := cachedData.(string)
// 					if !ok {
// 						msg := tgbotapi.NewMessage(u.SentFrom().ID, "Ошибка получения значения.")
// 						_, err := b.SendMessage(msg)
// 						return err
// 					}

// 					if err := config.Validator(value); err != nil {
// 						msg := tgbotapi.NewMessage(u.SentFrom().ID, "Значение некорректно: "+err.Error())
// 						_, sendErr := b.SendMessage(msg)
// 						if sendErr != nil {
// 							return sendErr
// 						}
// 						return err
// 					}

// 					if err := config.OnEnterAction(value, u.SentFrom().ID); err != nil {
// 						msg := tgbotapi.NewMessage(u.SentFrom().ID, "Ошибка обработки значения: "+err.Error())
// 						_, sendErr := b.SendMessage(msg)
// 						if sendErr != nil {
// 							return sendErr
// 						}
// 						return err
// 					}

// 					b.GetCache().Delete(cacheKey)

// 					msg := tgbotapi.NewMessage(u.SentFrom().ID, config.SuccessText)
// 					_, err := b.SendMessage(msg)
// 					if err != nil {
// 						return err
// 					}

// 					return b.SetUserState(u.SentFrom().ID, "")
// 				},
// 			},
// 		},
// 	}

// 	// Состояние для перехода в состояние слайдера
// 	var enterInState = tgfsm.State{
// 		Global: true,
// 	}
// 	var enterInStateID = uuid.New().String()

// 	if len(config.MessageTriggers) > 0 {
// 		enterInState.MessageHandlers = make(map[string]tgfsm.Handler)
// 		for _, t := range config.MessageTriggers {
// 			enterInState.MessageHandlers[t] = tgfsm.Handler{Handle: NewSetUserStateImmediateHandler(sliderStateID)}
// 		}
// 	}

// 	if len(config.CallbackTriggers) > 0 {
// 		enterInState.CallbackHandlers = make(map[string]tgfsm.Handler)
// 		for _, t := range config.CallbackTriggers {
// 			enterInState.CallbackHandlers[t] = tgfsm.Handler{Handle: NewSetUserStateImmediateHandler(sliderStateID)}
// 		}
// 	}

// 	return map[string]tgfsm.State{
// 		sliderStateID:  sliderPhase,
// 		enterInStateID: enterInState,
// 	}, nil
// }
