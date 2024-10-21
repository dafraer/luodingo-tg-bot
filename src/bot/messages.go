package bot

import (
	"flashcards-bot/src/db"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func processMessage(b *tgBot, update tgbotapi.Update) {
	//Check that message is not longer than maxMessageLen
	if len([]rune(update.Message.Text)) > maxMessageLen {
		longMessage(b, update)
	}
	b.Logger.Infow("Message", "from", update.Message.From.UserName, "body", update.Message.Text)

	user, err := db.GetUser(update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user state", "error", err.Error())
		return
	}
	switch user.State {
	case waitingNewDeckName:
		newDeckNameMessage(b, update)
	case waitingNewCardFront:
		newCardFrontMessage(b, update, user)
	case waitingNewCardBack:
		newCardBackMessage(b, update, user)
	default:
		unknownMessage(b, update)
	}

}

func longMessage(b *tgBot, update tgbotapi.Update) {
	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.TooLong[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

func newDeckNameMessage(b *tgBot, update tgbotapi.Update) {
	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//If deck exists already notify user about it
	//Decks cant have the same name
	exists, err := db.DeckExists(&db.Deck{Name: update.Message.Text, TgUserId: update.Message.From.ID})
	if err != nil {
		b.Logger.Errorw("Error checking if deck exists", "error", err.Error())
		return
	}

	if exists {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.DeckExists[lang])
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//Creating the deck in the database
	if err := db.CreateDeck(&db.Deck{Name: update.Message.Text, TgUserId: update.Message.From.ID}); err != nil {
		b.Logger.Errorw("Error creating deck", "error", err.Error())

		//If creating the deck failed - notify the user
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.ErrorCreatingDeck[lang])
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//Notify the user that deck has been created successfully
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.DeckCreated[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

func newCardFrontMessage(b *tgBot, update tgbotapi.Update, user *db.User) {
	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Update user state to "waiting for back side of the card" and put front card in there
	id, err := db.CreateCard(user.DeckSelected, user.TgUserId, &db.Card{Front: update.Message.Text, Back: "", Learned: false})
	if err != nil {
		b.Logger.Errorw("Error creating card", "error", err.Error())
	}

	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, State: waitingNewCardBack, CardSelected: id}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Prompt the user to choose back of the card
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.ChooseCardBack[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}

	b.clearDeleteQueue(update.Message.Chat.ID)
}

func newCardBackMessage(b *tgBot, update tgbotapi.Update, user *db.User) {
	if err := db.UpdateCard(&db.Card{Id: user.CardSelected, Back: update.Message.Text}); err != nil {
		b.Logger.Errorw("Error updating card", "error", err.Error())
	}

	//Update user state
	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, State: waitingNewCardFront, CardSelected: user.CardSelected}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Create an inline keyboard to stop adding cards or to add reverse card
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.AddReverse[lang], addReverse)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.Done[lang], done)),
	)

	//Prompt to add another card
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.ChooseCardFront[lang])
	msg.ReplyMarkup = keyboard
	sentMessage, err := b.Bot.Send(msg)
	if err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}

	//Add message to the delete queue to make sure that inline keyboard will be deleted later
	b.DeleteQueue = append(b.DeleteQueue, message{sentMessage.MessageID, sentMessage.Chat.ID})

}

func unknownMessage(b *tgBot, update tgbotapi.Update) {
	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.UnknownMessage[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}
