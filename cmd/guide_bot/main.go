package main

import (
	"fmt"
	"log"
	"tgfsm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

/*

 */
// main initializes and starts the guide bot with FSM states
// The bot shows user ID and file IDs for any media content sent to it
func main() {
	token := "YOUR_BOT_TOKEN"

	bot, err := tgfsm.NewBot(token, tgfsm.WithStates(States))
	if err != nil {
		log.Fatal(err)
	}

	bot.Start(0, 10)

	select {}
}

var States = map[string]tgfsm.State{
	"start": {
		Global: true,
		MessageHandlers: map[string]tgfsm.Handler{
			"/start": {Handle: HandleStart},
			"/help":  {Handle: HandleHelp},
			"/id":    {Handle: HandleID},
		},
	},
}

func HandleStart(b *tgfsm.Bot, u tgbotapi.Update) error {
	instruction := `🤖 <b>Guide Bot</b>

This bot demonstrates various functionalities:
• /help - Show help information
• /id - Show your user ID`

	msg := tgbotapi.NewMessage(u.Message.Chat.ID, instruction)
	msg.ParseMode = "HTML"
	_, err := b.SendMessage(msg)
	if err != nil {
		return err
	}

	return nil
}

func HandleHelp(b *tgfsm.Bot, u tgbotapi.Update) error {
	helpText := `📖 <b>Help</b>

<b>Commands:</b>
/start - Show instructions
/help - Show this help
/id - Show your user ID

<b>How to use:</b>
Just send me any message, sticker, photo, document, video, voice, audio, animation, video note, location, contact, or forward any message from another user/bot.

I'll show you all the relevant IDs!`

	msg := tgbotapi.NewMessage(u.Message.Chat.ID, helpText)
	msg.ParseMode = "HTML"
	_, err := b.SendMessage(msg)
	if err != nil {
		return err
	}
	return nil
}

func HandleID(b *tgfsm.Bot, u tgbotapi.Update) error {
	userID := u.Message.From.ID
	username := u.Message.From.UserName
	firstName := u.Message.From.FirstName
	lastName := u.Message.From.LastName

	msgText := fmt.Sprintf("👤 <b>Your User ID:</b> <code>%d</code>", userID)

	if username != "" {
		msgText += fmt.Sprintf("\n<b>Username:</b> @%s", username)
	}

	if firstName != "" {
		msgText += fmt.Sprintf("\n<b>Name:</b> %s", firstName)
		if lastName != "" {
			msgText += fmt.Sprintf(" %s", lastName)
		}
	}

	msg := tgbotapi.NewMessage(u.Message.Chat.ID, msgText)
	msg.ParseMode = "HTML"
	_, err := b.SendMessage(msg)
	if err != nil {
		return err
	}
	return nil
}
