package bot

import (
	"flashcards-bot/src/db"
	"flashcards-bot/src/text"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

var en, ru text.Messages

func (b *tgBot) Run() {
	en = text.LoadEnMessages()
	ru = text.LoadRuMessages()
	for update := range b.updates {
		handleUpdates(b, update)
	}
}

func handleUpdates(b *tgBot, update tgbotapi.Update) {
	switch {
	case update.Message != nil && update.Message.IsCommand():
		processCommand(b, update)
	case update.Message != nil:
		processMessage(b, update)
	case update.CallbackQuery != nil:
		processCallback(b, update)
	}
}

func createDecksInlineKeyboard(userId int64) (keyboard tgbotapi.InlineKeyboardMarkup, decksAmount int, err error) {
	//Get decks from database
	decks, err := db.GetDecks(userId)

	//Create buttons
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, v := range decks {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(v.Name), fmt.Sprint(v.Name))))
	}

	//Return the keyboard with created buttons
	return tgbotapi.NewInlineKeyboardMarkup(buttons...), len(decks), err
}

func createCardsInlineKeyboard(userId int64, deckName string) (keyboard tgbotapi.InlineKeyboardMarkup, cardsAmount int, err error) {
	//Get cards from database
	cards, err := db.GetCards(deckName, userId)

	//Create buttons with front-back of a card shown to the user and card id sent as a callback data
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, v := range cards {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s-%s", v.Front, v.Back), fmt.Sprint(v.Id))))
	}

	//Return the keyboard with created buttons
	return tgbotapi.NewInlineKeyboardMarkup(buttons...), len(cards), err
}

func (b *tgBot) deleteMessage(chatId int64, messageId int) {
	deleteMessage := tgbotapi.NewDeleteMessage(
		chatId,
		messageId,
	)
	if _, err := b.bot.Send(deleteMessage); err != nil {
		log.Printf("Error deleting message: %v\n", err)
	}
}
