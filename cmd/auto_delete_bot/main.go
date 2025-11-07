package main

import (
	"fmt"
	"log"
	"tgfsm"
	"tgfsm/cache"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*
Simple echo bot demonstrating auto-deletion feature with go-cache.
Each new message automatically deletes the previous non-important message.
*/
func main() {

	token := "YOUR_BOT_TOKEN"

	// Create go-cache implementation for last messages
	lastMsgCache := cache.NewGoCacheLastMessage(
		30*24*time.Hour, // expiration - 30 days
		1*time.Hour,     // cleanupInterval - 1 hour
	)

	// Create bot with auto-deletion enabled
	bot, err := tgfsm.NewBot(token,
		tgfsm.WithAutoDelete(true),
		tgfsm.WithLastMessageCache(lastMsgCache),
		tgfsm.WithUpdateHandler(Echo),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Bot started! Send any text message to see auto-deletion in action.")
	bot.Start(0, 10)

	select {}
}

// Echo echoes the message back to the user
// Each new message will automatically delete the previous one
// Send "important" to see an important message that won't be deleted
func Echo(b *tgfsm.Bot, u tgbotapi.Update) error {
	if u.Message != nil && u.Message.Text != "" {
		text := u.Message.Text

		// If user sends "important", send an important message that won't be deleted
		if text == "important" {
			importantMsg := tgbotapi.NewMessage(u.Message.Chat.ID, "This is an important message! It won't be deleted automatically.")
			_, err := b.SendImportantMessage(importantMsg)
			if err != nil {
				fmt.Println("Error sending important message:", err)
				return err
			}
			return nil
		}

		// Simple echo - bot repeats the message
		// This message will be deleted when next message is sent
		msg := tgbotapi.NewMessage(u.Message.Chat.ID, text)
		_, err := b.SendMessage(msg)
		if err != nil {
			fmt.Println("Error sending message:", err)
			return err
		}
	}
	return nil
}
