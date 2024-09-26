package bot

import (
	"flashcards-bot/src/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math/rand"
)

func (b *tgBot) Run() {
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

func createDecksInlineKeyboard(userId int64, page int) (keyboard tgbotapi.InlineKeyboardMarkup, decksAmount int, err error) {
	//Get decks from database
	decks, err := db.GetDecks(userId)

	//Figure out from which card to display
	from := (page - 1) * 10

	//Create buttons
	var buttons [][]tgbotapi.InlineKeyboardButton
	for i := from; i < min(len(decks), from+10); i++ {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(decks[i].Name), fmt.Sprint(decks[i].Name))))
	}

	//add change page button
	if len(decks) >= 11 {
		switch {
		case from == 0:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➡️️", rightDeck)))
		case from >= len(decks)-10:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", leftDeck)))
		default:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", leftDeck), tgbotapi.NewInlineKeyboardButtonData("➡️️", rightDeck)))
		}
	}
	//Return the keyboard with created buttons
	return tgbotapi.NewInlineKeyboardMarkup(buttons...), len(decks), err
}

func createCardsInlineKeyboard(userId int64, deckName string, b *tgBot, page int) (keyboard tgbotapi.InlineKeyboardMarkup, cardsAmount int, err error) {
	b.Logger.Debugw("createCardsInlineKeyboard", "page", page)
	//Get cards from database
	cards, err := db.GetCards(deckName, userId)

	//Figure out from which card to display
	from := (page - 1) * 10

	//Create buttons with front-back of a card shown to the user and card id sent as a callback data
	var buttons [][]tgbotapi.InlineKeyboardButton
	for i := from; i < min(from+10, len(cards)); i++ {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s-%s", cards[i].Front, cards[i].Back), fmt.Sprint(cards[i].Id))))
	}
	//add change page button
	if len(cards) >= 10 {
		switch {
		case from == 0:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➡️️", rightCard)))
		case from >= len(cards)-10:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", leftCard)))
		default:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", leftCard), tgbotapi.NewInlineKeyboardButtonData("➡️️", rightCard)))
		}
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
		//Avoiding error handling on purpose
		//Because it does not matter if we cant delete message
		b.Bot.Request(deleteMessage)
	}
	b.DeleteQueue = nil
}

// Creates a message with a card to study
func (b *tgBot) studyRandomCard(update tgbotapi.Update) (tgbotapi.EditMessageTextConfig, error) {
	//Get user state to know what to know selected deck
	user, err := db.GetUser(update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user state", "error", err.Error())
		return tgbotapi.EditMessageTextConfig{}, err
	}

	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
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
			b.Messages.FinishedStudy[lang],
		)
		return edit, nil
	}

	//Pick a random card
	card := cards[rand.Intn(len(cards))]

	//Create buttons with 2 options:
	//Show back of the card
	//Stop studying
	var buttons [][]tgbotapi.InlineKeyboardButton
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.ShowAnswer[lang], fmt.Sprint(card.Back))))
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.StopStudy[lang], stop)))

	//Created an inline keyboard with previously created buttons
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	//Edit already existing message to remove previous inline keyboard
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		card.Front+"\n——————————————————————",
		keyboard,
	)

	user.CardSelected = card.Id
	if err := db.UpdateUser(user); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}
	return edit, nil
}

func language(languageCode string, userId int64) (string, error) {
	if languageCode != "en" && languageCode != "ru" && languageCode != "es" {
		user, err := db.GetUser(userId)
		if err != nil {
			return "", err
		}

		if user.Language != "en" && user.Language != "ru" && user.Language != "es" {
			return "en", nil
		}
		return user.Language, nil
	}
	return languageCode, nil
}
