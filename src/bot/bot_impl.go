package bot

import (
	"flashcards-bot/src/db"
	"flashcards-bot/src/text"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func createDecksInlineKeyboard(b *tgBot, update tgbotapi.Update) (keyboard tgbotapi.InlineKeyboardMarkup, decksAmount int, err error) {
	//Get decks from database
	decks, err := db.GetDecks(update.Message.From.ID)

	//Create buttons
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, v := range decks {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(v.Name), fmt.Sprint(v.Name))))
	}

	//Return the keyboard with created buttons
	return tgbotapi.NewInlineKeyboardMarkup(buttons...), len(decks), err
}

//*******************************
// REFACTOR EVERYTHING UNDER THIS COMMENT
//*******************************

func listCardsHandler(b *tgBot, update tgbotapi.Update) {

}
