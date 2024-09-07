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
	case myCards:
		listCardsHandler(b, update)
	case deleteDeck:
		deleteDeckHandler(b, update)
	case deckDeleteCard:
		deleteCardHandler(b, update)
	case cardDeleteCard:
		deleteCardHandler(b, update)
	case deckNewCard:
		newCardHandler(b, update)
	case cardNewCard:
		newCardHandler(b, update)
	case studyDeck:
		studyDeckHandler(b, update)
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
		if _, err := b.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
		return
	}

	//Notify the user that deck has been created successfully
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.DeckCreated)
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func unknownMessage(b *tgBot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.UnknownMessage)
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}
