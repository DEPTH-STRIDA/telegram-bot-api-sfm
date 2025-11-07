package tgfsm

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) SendDeleteMessage(msg tgbotapi.DeleteMessageConfig) (*tgbotapi.APIResponse, error) {
	b.limiter.WaitForAPI(context.Background())

	sendedMsg, err := b.BotAPI.Request(msg)
	if err != nil {
		return nil, err
	}

	return sendedMsg, nil
}

func (b *Bot) SendMessage(msg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	// Auto-delete last message if enabled
	if b.autoDeleteEnabled && b.lastMessageCache != nil {
		lastMsgInfo, err := b.lastMessageCache.GetLastMessageInfo(msg.ChatID)
		if err == nil && lastMsgInfo != nil && lastMsgInfo.MessageID > 0 {
			// Delete only if last message was not important
			if !lastMsgInfo.Important {
				// Try to delete last message (ignore errors)
				deleteMsg := tgbotapi.NewDeleteMessage(msg.ChatID, lastMsgInfo.MessageID)
				_ = b.DeleteMessage(deleteMsg)
			}
		}
	}

	b.limiter.WaitForMessage(context.Background(), msg.ChatID)

	sendedMsg, err := b.BotAPI.Send(msg)
	if err != nil {
		return sendedMsg, err
	}

	// Save information about sent message to cache if auto-deletion is enabled
	if b.autoDeleteEnabled && b.lastMessageCache != nil {
		isImportant := isImportantMessage(msg)
		info := &LastMessageInfo{
			MessageID: sendedMsg.MessageID,
			Important: isImportant,
		}
		_ = b.lastMessageCache.SetLastMessageInfo(msg.ChatID, info)
	}

	return sendedMsg, nil
}

// isImportantMessage determines if a message is important
// Important messages are not automatically deleted
// A message is considered important if it has ReplyMarkup (keyboard)
func isImportantMessage(msg tgbotapi.MessageConfig) bool {
	return msg.ReplyMarkup != nil
}

// SendImportantMessage sends a message and marks it as important
// Important messages are not automatically deleted during auto-deletion
// This method doesn't require ReplyMarkup - message is marked as important explicitly
func (b *Bot) SendImportantMessage(msg tgbotapi.MessageConfig) (tgbotapi.Message, error) {
	// Auto-delete last message if enabled
	if b.autoDeleteEnabled && b.lastMessageCache != nil {
		lastMsgInfo, err := b.lastMessageCache.GetLastMessageInfo(msg.ChatID)
		if err == nil && lastMsgInfo != nil && lastMsgInfo.MessageID > 0 {
			// Delete only if last message was not important
			if !lastMsgInfo.Important {
				// Try to delete last message (ignore errors)
				deleteMsg := tgbotapi.NewDeleteMessage(msg.ChatID, lastMsgInfo.MessageID)
				_ = b.DeleteMessage(deleteMsg)
			}
		}
	}

	b.limiter.WaitForMessage(context.Background(), msg.ChatID)

	sendedMsg, err := b.BotAPI.Send(msg)
	if err != nil {
		return sendedMsg, err
	}

	// Save information about sent message to cache if auto-deletion is enabled
	// Mark as important explicitly
	if b.autoDeleteEnabled && b.lastMessageCache != nil {
		info := &LastMessageInfo{
			MessageID: sendedMsg.MessageID,
			Important: true, // Explicitly mark as important
		}
		_ = b.lastMessageCache.SetLastMessageInfo(msg.ChatID, info)
	}

	return sendedMsg, nil
}

// NewImportantMessage creates a message marked as important (deprecated, use SendImportantMessage instead)
// Important messages are not automatically deleted during auto-deletion
// Uses inline keyboard with invisible button to mark message as important
// Deprecated: Use SendImportantMessage method instead - it doesn't require keyboard
func NewImportantMessage(chatID int64, text string) tgbotapi.MessageConfig {
	msg := tgbotapi.NewMessage(chatID, text)
	// Use inline keyboard with invisible button to mark as important
	// Button uses zero-width space character to be invisible
	// This marks message as important so it won't be deleted
	invisibleButton := tgbotapi.NewInlineKeyboardButtonData("\u200B", "important")
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(invisibleButton),
	)
	msg.ReplyMarkup = keyboard
	return msg
}

func (b *Bot) SendPinMessageEvent(messageID int, ChatID int64, disableNotification bool) (*tgbotapi.APIResponse, error) {
	b.limiter.WaitForAPI(context.Background())

	pinConfig := tgbotapi.PinChatMessageConfig{
		ChatID:              ChatID,
		MessageID:           messageID,
		DisableNotification: disableNotification,
	}

	APIResponse, err := b.BotAPI.Request(pinConfig)
	if err != nil {
		return nil, err
	}
	return APIResponse, nil
}

func (b *Bot) SendSticker(stickerID string, chatID int64) (*tgbotapi.Message, error) {
	b.limiter.WaitForMessage(context.Background(), chatID)

	msg := tgbotapi.NewSticker(chatID, tgbotapi.FileID(stickerID))

	sendedMsg, err := b.BotAPI.Send(msg)
	if err != nil {
		return nil, err
	}

	return &sendedMsg, nil
}

func (b *Bot) SendUnPinAllMessageEvent(ChannelUsername string, chatID int64) (*tgbotapi.APIResponse, error) {
	b.limiter.WaitForAPI(context.Background())

	unpinConfig := tgbotapi.UnpinAllChatMessagesConfig{
		ChatID:          chatID,
		ChannelUsername: ChannelUsername,
	}

	APIresponse, err := b.BotAPI.Request(unpinConfig)
	if err != nil {
		return nil, err
	}

	return APIresponse, err
}

func (b *Bot) EditMessage(editMsg tgbotapi.EditMessageTextConfig) (*tgbotapi.APIResponse, error) {
	b.limiter.WaitForAPI(context.Background())

	response, err := b.BotAPI.Request(editMsg)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (b *Bot) DeleteMessage(deleteMsg tgbotapi.DeleteMessageConfig) error {
	b.limiter.WaitForAPI(context.Background())

	_, err := b.BotAPI.Request(deleteMsg)
	if err != nil {
		return err
	}

	return err
}
