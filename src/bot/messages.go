package bot

import (
	"flashcards-bot/src/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func processMessage(b *tgBot, update tgbotapi.Update) {
	log.Printf("MESSAGE: [%s] %s\n", update.Message.From.UserName, update.Message.Text)
	user, err := db.GetUserState(update.Message.From.ID)
	if err != nil {
		log.Printf("Error getting user state: %v\n", err)
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
	if err := db.CreateDeck(update.Message.Text, update.Message.From.ID); err != nil {
		log.Printf("Error creating deck: %v\n", err)

		//If creating the deck failed - notify the user
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ErrorCreatingDeck)
		if _, err := b.Bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
		return
	}

	//Notify the user that deck has been created successfully
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.DeckCreated)
	if _, err := b.Bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func newCardFrontMessage(b *tgBot, update tgbotapi.Update) {

	//Update user state to "waiting for back side of the card" and put front card in there
	if err := db.UpdateUserState(db.User{TgUserId: update.Message.From.ID, State: waitingNewCardBack, CardSelected: update.Message.Text}); err != nil {
		log.Printf("Error updating user state: %v\n", err)
	}

	//Prompt the user to choose back of the card
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseCardBack)
	if _, err := b.Bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func newCardBackMessage(b *tgBot, update tgbotapi.Update) {
	//Get user data
	user, err := db.GetUserState(update.Message.From.ID)
	if err != nil {
		log.Printf("Error getting user state: %v\n", err)
	}

	//Create the card in the db
	if err := db.CreateCard(user.DeckSelected, user.TgUserId, user.CardSelected, update.Message.Text); err != nil {
		log.Printf("Error creating deck: %v\n", err)
	}

	//Update user state
	if err := db.UpdateUserState(db.User{TgUserId: update.Message.From.ID, State: defaultState, DeckSelected: " ", CardSelected: " "}); err != nil {
		log.Printf("Error updating user state: %v\n", err)
	}

	//Notify the user about creating card
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.CardCreated)
	if _, err := b.Bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}

}

func unknownMessage(b *tgBot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.UnknownMessage)
	if _, err := b.Bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}
