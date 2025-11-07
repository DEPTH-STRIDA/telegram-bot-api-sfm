package main

import (
	"fmt"
	"log"
	"tgfsm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*
A simple bot that responds to a user's chat with the same message. Only for text messages.
*/
// main initializes and starts the echo bot
// The bot simply repeats back any text message sent to it
func main() {
	token := "YOUR_BOT_TOKEN"

	bot, err := tgfsm.NewBot(token, tgfsm.WithUpdateHandler(Echo))
	if err != nil {
		log.Fatal(err)
	}

	bot.Start(0, 10)

	select {}
}

// Echo is a function that echoes the message back to the user
func Echo(b *tgfsm.Bot, u tgbotapi.Update) error {

	if u.Message != nil {
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, u.Message.Text)
		_, err := b.SendMessage(msg)
		if err != nil {
			fmt.Println("Error sending message: ", err)
			return err
		}
	}
	return nil
}
