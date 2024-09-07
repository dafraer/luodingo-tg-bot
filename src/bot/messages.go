package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func processMessage(b *tgBot, update tgbotapi.Update) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	b.sessions.mutex.RLock()
	state := b.sessions.userStates[update.Message.From.ID].action
	b.sessions.mutex.RUnlock()
	switch state {
	case newDeck:
		newDeckHandler(b, update)
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
		unknownMessageHandler(b, update)
	}

}
