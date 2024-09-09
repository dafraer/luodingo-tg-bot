// This file contains functions that process commands
// Commands generally do not depend on user state, but they update it for future messages and callback queries
package bot

import (
	"flashcards-bot/src/db"
	"fmt"
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
	//Create new user if it does not exist
	_, err := db.GetUserState(update.Message.From.ID)
	if err != nil {
		//TODO fix
		log.Printf("Error getting user state: %s\n", err)
		if err := db.CreateUser(db.User{TgUserId: update.Message.From.ID}); err != nil {
			log.Printf("Error creating user: %v", err)
		}
	}

	//Send start message
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
	//Update user state to "waiting for a deck in which card should be created" and put deck name in there
	if err := db.UpdateUserState(db.User{TgUserId: update.Message.From.ID, State: waitingNewCardDeckName}); err != nil {
		log.Printf("Error updating user state: %v\n", err)
	}

	//Create a message
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseDeck)

	//Create an inline keyboard
	keyboard, decksAmount, err := createDecksInlineKeyboard(update.Message.From.ID)
	if err != nil {
		log.Printf("Error creating a keyboard: %v\n", err)
	}

	//If user has no decks prompt him to create one first
	if decksAmount <= 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.CreateDeckFirst)
		if _, err := b.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
		return
	}

	// Attaching the keyboard to the message
	msg.ReplyMarkup = keyboard

	// Sending the message with the attached inline keyboard
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}
func listDecksCommand(b *tgBot, update tgbotapi.Update) {
	//Get decks from db
	decks, err := db.GetDecks(update.Message.From.ID)
	if err != nil {
		log.Printf("Error getting decks from db: %v", err)
	}

	//If no decks tell user that they have no decks
	if len(decks) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.CreateDeckFirst)
		if _, err := b.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
		return
	}

	//List user's decks
	table := "These are your decks:\n"
	for i, v := range decks {
		table += fmt.Sprintf("%d. %v\n", i+1, v)
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, table)
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func listCardsCommand(b *tgBot, update tgbotapi.Update) {
	//Update user state to "waiting for a deck name to delete"
	if err := db.UpdateUserState(db.User{TgUserId: update.Message.From.ID, State: waitingListMyCardsDeckName}); err != nil {
		log.Printf("Error updating user state: %v\n", err)
		return
	}

	//Create a keyboard with decks
	keyboard, decksAmount, err := createDecksInlineKeyboard(update.Message.From.ID)
	if err != nil {
		log.Printf("Error creating a keyboard: %v\n", err)
		return
	}

	//If user has no decks tell them that
	if decksAmount <= 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.NoDecks)
		if _, err := b.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
	}

	//Prompt user to choose deck
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseDeckName)
	msg.ReplyMarkup = keyboard
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func deleteDeckCommand(b *tgBot, update tgbotapi.Update) {
	//Update user state to "waiting for a deck name to delete"
	if err := db.UpdateUserState(db.User{TgUserId: update.Message.From.ID, State: waitingDeleteDeckName}); err != nil {
		log.Printf("Error updating user state: %v\n", err)
	}

	//Create a new message
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseDeck)

	//Create a new keyboard with decks to choose from
	keyboard, decksAmount, err := createDecksInlineKeyboard(update.Message.From.ID)
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
	//Update user state to "waiting for a deck to delete a card from"
	if err := db.UpdateUserState(db.User{TgUserId: update.Message.From.ID, State: waitingDeleteCardDeckName}); err != nil {
		log.Printf("Error updating user state: %v\n", err)
	}

	//Create  a keyboard with decks to choose from
	keyboard, decksAmount, err := createDecksInlineKeyboard(update.Message.From.ID)
	if err != nil {
		log.Printf("Error creating inline keyboard:%v\n", err)
	}

	//If user has no decks tell them that
	if decksAmount <= 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.NoDecks)
		if _, err := b.bot.Send(msg); err != nil {
			log.Printf("Error sending message: %v\n", err)
		}
		return
	}

	//Create and send the message
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseDeck)
	msg.ReplyMarkup = keyboard
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func studyDeckCommand(b *tgBot, update tgbotapi.Update) {

}

func unknownCommand(b *tgBot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.UnknownCommand)
	if _, err := b.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}
