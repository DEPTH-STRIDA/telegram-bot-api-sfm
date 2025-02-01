### Пример использования бота

```go
package main

import (
	"app/states"
    "os"
	"time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	tgbot "github.com/depth-strida/telegram-bot-api-sfm"
)

func main() {
	// Создаем конфигурацию бота
	config := tgbot.Config{
		Token:           os.Getenv("TELEGRAM_APITOKEN"),
		Expiration:      24 * time.Hour, // Время жизни состояний
		CleanupInterval: time.Hour,      // Интервал очистки
		States:          states,
	}

	b, err := tgbot.NewBot(config)
	if err != nil {
		panic(err)
	}

	b.Start(0, 30)

	select {}
}

var states = map[string]tgbot.State{
	"start": {
		Global: true,
		MessageHandlers: map[string]tgbot.Handler{
			"/start": {
				Handle: func(b *tgbot.Bot, u tgbotapi.Update) error {
                    // Подготовка конфига сообщения
					msg := tgbotapi.NewMessage(u.SentFrom().ID, "Привет и тебе")
                    // Отправка через метод, чтобы ограничитель учитывал отправку сообщения
					_, err := b.SendMessage(msg)
					if err != nil {
						fmt.Println(err)
					}
					return nil
				},
			},
		},
	},
}
```