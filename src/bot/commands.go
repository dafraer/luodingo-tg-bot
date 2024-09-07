// This file contains functions that process commands
// Commands generally do not depend on user state, but they update it for future messages and callback queries
package bot

import (
	"flashcards-bot/src/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func processCommand(b *tgBot, update tgbotapi.Update) {
	log.Printf("COMMAND: [%s] %s\n", update.Message.From.UserName, update.Message.Text)
	switch update.Message.Command() {
	case "start":
		startCommand(b, update)
	case "help":
		helpCommand(b, update)
	case "new_deck":
		newDeckCommand(b, update)
	case "new_card":
		newCardCommand(b, update)
	case "my_cards":
		listCardsCommand(b, update)
	case "my_decks":
		listDecksCommand(b, update)
	case "delete_deck":
		deleteDeckCommand(b, update)
	case "delete_card":
		deleteCardCommand(b, update)
	case "study_deck":
		studyDeckCommand(b, update)
	default:
		unknownCommand(b, update)
	}
}

func startCommand(b *tgBot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.Start)
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func helpCommand(b *tgBot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.Help)
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func newDeckCommand(b *tgBot, update tgbotapi.Update) {
	//Update user state to "waiting for a deck name to create new deck"
	if err := db.UpdateUserState(db.User{TgUserId: update.Message.From.ID, State: waitingNewDeckName}); err != nil {
		log.Printf("Error updating user state: %v\n", err)
	}

	//Send the next message to the user
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseDeckName)
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}
func newCardCommand(b *tgBot, update tgbotapi.Update) {

}
func listDecksCommand(b *tgBot, update tgbotapi.Update) {

}

func listCardsCommand(b *tgBot, update tgbotapi.Update) {

}

func deleteDeckCommand(b *tgBot, update tgbotapi.Update) {
	//Update user state to "waiting for a deck name to delete"
	if err := db.UpdateUserState(db.User{TgUserId: update.Message.From.ID, State: waitingDeleteDeckName}); err != nil {
		log.Printf("Error updating user state: %v\n", err)
	}

	//Create a new message
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseDeck)

	//Create a new keyboard with decks to choose from
	keyboard, decksAmount, err := createDecksInlineKeyboard(b, update)
	if err != nil {
		log.Printf("Error getting decks:%v\n", err)
	}

	//If user has no decks let them now
	if decksAmount == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.NoDecks)
		if _, err := b.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
		return
	}

	// Sending the message with the attached inline keyboard
	msg.ReplyMarkup = keyboard
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func deleteCardCommand(b *tgBot, update tgbotapi.Update) {

}

func studyDeckCommand(b *tgBot, update tgbotapi.Update) {

}

func unknownCommand(b *tgBot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.UnknownCommand)
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}
