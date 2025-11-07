package main

import (
	"fmt"
	"log"
	"strings"
	"tgfsm"
	"tgfsm/events"
	"unicode/utf8"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Chat ID where feedback will be sent
const CHAT_ID = 000000000

// main initializes and starts the feedback bot
// The bot collects user feedback with name and message
func main() {
	token := "YOUR_BOT_TOKEN"

	var bot *tgfsm.Bot
	botPtr := &bot

	// Создаем состояния для сбора имени используя опции
	states, err := events.NewEnterDataEvent(
		events.WithMessageTriggers("/start", "/feedback", "отзыв"),
		events.WithValidator(validateName),
		events.WithSubmitText("Подтвердить"),
		events.WithOnSuccessEnterAction(handleNameEntered(botPtr)),
		events.WithPromptText("Как Вас зовут?"),
		events.WithConfirmInputText("Вы ввели:\n\n%s\n\nДля подтверждения нажмите кнопку ниже. В случае ошибки введите данные заново."),
		events.WithSuccessText("Спасибо! Ваше имя сохранено."),
		events.WithErrorText("Произошла ошибка при сохранении имени."),
		events.WithValidationErrorText("Некорректные данные: %s"),
	)
	if err != nil {
		log.Fatal(err)
	}

	bot, err = tgfsm.NewBot(token, tgfsm.WithStates(states))
	if err != nil {
		log.Fatal(err)
	}

	bot.Start(0, 10)

	select {}
}

// validateName проверяет корректность введенного имени
func validateName(name string) error {
	name = strings.TrimSpace(name)
	runeCount := utf8.RuneCountInString(name)
	if runeCount < 2 {
		return fmt.Errorf("имя должно содержать минимум 2 символа")
	}
	if runeCount > 50 {
		return fmt.Errorf("имя не должно превышать 50 символов")
	}
	return nil
}

// handleNameEntered обрабатывает введенное имя
func handleNameEntered(botPtr **tgfsm.Bot) func(message string, userId int64) error {
	return func(message string, userId int64) error {
		if *botPtr == nil {
			return fmt.Errorf("bot is not initialized")
		}

		msg := fmt.Sprintf(`=== ОТЗЫВ ===
Пользователь ID: %d
Имя: %s
==============`, userId, message)

		fmt.Println(msg)

		_, err := (*botPtr).SendMessage(tgbotapi.NewMessage(CHAT_ID, msg))

		return err
	}
}
