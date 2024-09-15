package bot

import (
	"flashcards-bot/src/db"
	"flashcards-bot/src/text"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math/rand"
)

var en, ru text.Messages

func (b *tgBot) Run() {
	en = text.LoadEnMessages()
	ru = text.LoadRuMessages()
	for update := range b.Updates {
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

func createCardsInlineKeyboard(userId int64, deckName string, b *tgBot) (keyboard tgbotapi.InlineKeyboardMarkup, cardsAmount int, err error) {
	//Get cards from database
	cards, err := db.GetCards(deckName, userId)

	b.Logger.Debugw("Got cards list", "cards", cards, "error", err)

	//Create buttons with front-back of a card shown to the user and card id sent as a callback data
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, v := range cards {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s-%s", v.Front, v.Back), fmt.Sprint(v.Id))))
		b.Logger.Debugw("Created new button", "message", fmt.Sprintf("%s-%s", v.Front, v.Back), "data", fmt.Sprint(v.Id))
	}

	//Return the keyboard with created buttons
	return tgbotapi.NewInlineKeyboardMarkup(buttons...), len(cards), err
}

func (b *tgBot) deleteMessage(chatId int64, messageId int) {
	//Delete message
	deleteMessage := tgbotapi.NewDeleteMessage(chatId, messageId)
	if _, err := b.Bot.Request(deleteMessage); err != nil {
		b.Logger.Errorw("Error deleting message", "error", err.Error())
	}
	b.Logger.Debugw("Delete message queue", "chatId", chatId, "messageId", messageId, "len", len(b.DeleteQueue))

	//Clear delete queue
	for _, msgId := range b.DeleteQueue {
		deleteMessage := tgbotapi.NewDeleteMessage(chatId, msgId)
		if _, err := b.Bot.Request(deleteMessage); err != nil {
			b.Logger.Errorw("Error deleting message", "error", err.Error())
		}
	}
}

// Creates a message with a card to study
func (b *tgBot) studyRandomCard(update tgbotapi.Update) (tgbotapi.EditMessageTextConfig, error) {
	//Get user state to know what to know selected deck
	user, err := db.GetUser(update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user state", "error", err.Error())
		return tgbotapi.EditMessageTextConfig{}, err
	}

	//Get cards from the selected deck
	cards, err := db.GetUnlearnedCards(user.DeckSelected, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting cards", "error", err.Error())
	}

	//If not enough cards tell the user
	if len(cards) == 0 {
		if err := db.UnlearnCards(user.DeckSelected, update.CallbackQuery.From.ID); err != nil {
			b.Logger.Errorw("Error unlearning cards", "error", err.Error())
		}

		edit := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			en.FinishedStudy,
		)
		return edit, nil
	}

	//Pick a random card
	card := cards[rand.Intn(len(cards))]

	//Create buttons with 2 options:
	//Show back of the card
	//Stop studying
	var buttons [][]tgbotapi.InlineKeyboardButton
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(en.ShowAnswer, fmt.Sprint(card.Back))))
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(en.StopStudy, "stop")))

	//Created an inline keyboard with previously created buttons
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	//Edit already existing message to remove previous inline keyboard
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		card.Front+"\n————————————————————",
		keyboard,
	)

	user.CardSelected = card.Front
	if err := db.UpdateUser(user); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}
	return edit, nil
}
