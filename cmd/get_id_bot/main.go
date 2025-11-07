package main

import (
	"fmt"
	"log"
	"tgfsm"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// main initializes and starts the ID bot
// The bot shows user ID and file IDs for any media content sent to it
func main() {
	token := "YOUR_BOT_TOKEN"

	bot, err := tgfsm.NewBot(token, tgfsm.WithUpdateHandler(ShowID))
	if err != nil {
		log.Fatal(err)
	}

	bot.Start(0, 10)

	select {}
}

func ShowID(b *tgfsm.Bot, u tgbotapi.Update) error {
	if u.Message == nil {
		return nil
	}

	msgText := fmt.Sprintf("Your ID: <code>%d</code>", u.Message.From.ID)

	switch {
	case u.Message.Sticker != nil:
		msgText = fmt.Sprintf(msgText+"\n\nSticker: <code>%s</code>", u.Message.Sticker.FileID)
	case u.Message.Photo != nil:
		msgText = fmt.Sprintf(msgText+"\n\nPhoto: <code>%s</code>", u.Message.Photo[0].FileID)
	case u.Message.Document != nil:
		msgText = fmt.Sprintf(msgText+"\n\nDocument: <code>%s</code>", u.Message.Document.FileID)
	case u.Message.Video != nil:
		msgText = fmt.Sprintf(msgText+"\n\nVideo: <code>%s</code>", u.Message.Video.FileID)
	case u.Message.Voice != nil:
		msgText = fmt.Sprintf(msgText+"\n\nVoice: <code>%s</code>", u.Message.Voice.FileID)
	case u.Message.Audio != nil:
		msgText = fmt.Sprintf(msgText+"\n\nAudio: <code>%s</code>", u.Message.Audio.FileID)
	case u.Message.Animation != nil:
		msgText = fmt.Sprintf(msgText+"\n\nAnimation: <code>%s</code>", u.Message.Animation.FileID)
	case u.Message.VideoNote != nil:
		msgText = fmt.Sprintf(msgText+"\n\nVideoNote: <code>%s</code>", u.Message.VideoNote.FileID)
	}

	msg := tgbotapi.NewMessage(u.Message.Chat.ID, msgText)
	msg.ParseMode = "HTML"
	_, err := b.SendMessage(msg)
	if err != nil {
		return err
	}

	return tgfsm.ErrStatesNil

}
