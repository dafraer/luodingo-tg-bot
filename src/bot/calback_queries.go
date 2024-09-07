package bot

import (
	"flashcards-bot/src/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func processCallback(b *tgBot, update tgbotapi.Update) {
	log.Printf("CALLBACK: [%s] %s\n", update.CallbackQuery.From.UserName, update.CallbackQuery.Data)
	user, err := db.GetUserState(update.CallbackQuery.From.ID)
	if err != nil {
		log.Printf("Error getting user state: %v\n", err)
	}
	switch user.State {
	case waitingDeleteDeckName:
		deleteDeckCallback(b, update)
	case waitingNewCardDeckName:
		newCardCallback(b, update)
	default:
		unknownCallback(b, update)
	}
}

func deleteDeckCallback(b *tgBot, update tgbotapi.Update) {
	//Update user state to default state
	/*if err := db.UpdateUserState(db.User{TgUserId: update.CallbackQuery.From.ID, State: defaultState}); err != nil {
		log.Printf("Error updating user state: %v\n", err)
	}
	*/
	//Get deck name to delete
	name := update.CallbackQuery.Data

	//Delete the deck
	if err := db.DeleteDeck(name, update.CallbackQuery.From.ID); err != nil {
		log.Printf("Error deleting deck: %v\n", err)

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.ErrorDeletingDeck)
		if _, err := b.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
	}

	//Get decks keyboard
	keyboard, decksAmount, err := createDecksInlineKeyboard(update.CallbackQuery.From.ID)
	if err != nil {
		log.Printf("Error getting inline keyboard: %v\n", err)
	}
	//if user has no decks left delete the message
	if decksAmount <= 0 {
		deleteMessage := tgbotapi.NewDeleteMessage(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
		)
		if _, err := b.bot.Send(deleteMessage); err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.DeckDeleted)
		if _, err := b.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
		return
	}

	//Update inline keyboard
	edit := tgbotapi.NewEditMessageReplyMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		keyboard,
	)
	if _, err := b.bot.Send(edit); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.DeckDeleted)
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func unknownCallback(b *tgBot, update tgbotapi.Update) {
	log.Printf("UNKNOWN CALLBACK: [%s]\n", update.CallbackQuery.From.UserName)
}

func newCardCallback(b *tgBot, update tgbotapi.Update) {
	//Update the user state to "waiting card front"
	if err := db.UpdateUserState(db.User{TgUserId: update.CallbackQuery.From.ID, State: waitingNewCardFront, DeckSelected: update.CallbackQuery.Data}); err != nil {
	}
	//Edit the message
	edit := tgbotapi.NewEditMessageText(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		en.ChooseCardFront,
	)

	//Send the edit
	if _, err := b.bot.Send(edit); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}
