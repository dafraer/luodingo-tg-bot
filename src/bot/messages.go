package bot

import (
	"flashcards-bot/src/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func processMessage(b *tgBot, update tgbotapi.Update) {
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
		newCardFrontMessage(b, update)
	case waitingNewCardBack:
		newCardBackMessage(b, update)
	default:
		unknownMessage(b, update)
	}

}

func newDeckNameMessage(b *tgBot, update tgbotapi.Update) {
	//Creating the deck in the database
	if err := db.CreateDeck(&db.Deck{Name: update.Message.Text, TgUserId: update.Message.From.ID, CardsAmount: 0}); err != nil {
		b.Logger.Errorw("Error creating deck", "error", err.Error())

		//If creating the deck failed - notify the user
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ErrorCreatingDeck)
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//Notify the user that deck has been created successfully
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.DeckCreated)
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

func newCardFrontMessage(b *tgBot, update tgbotapi.Update) {

	//Update user state to "waiting for back side of the card" and put front card in there
	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, State: waitingNewCardBack, CardSelected: update.Message.Text}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Prompt the user to choose back of the card
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseCardBack)
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

func newCardBackMessage(b *tgBot, update tgbotapi.Update) {
	//Get user data
	user, err := db.GetUser(update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user state", "error", err.Error())
	}

	//Create the card in the db
	if err := db.CreateCard(user.DeckSelected, user.TgUserId, &db.Card{Front: user.CardSelected, Back: update.Message.Text}); err != nil {
		b.Logger.Errorw("Error creating deck", "error", err.Error())
	}

	//Update user state
	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, State: defaultState, DeckSelected: " ", CardSelected: " "}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Notify the user about creating card
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.CardCreated)
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}

}

func unknownMessage(b *tgBot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.UnknownMessage)
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}
