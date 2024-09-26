// This file contains functions that process commands
// Commands generally do not depend on user state, but they update it for future messages and callback queries
package bot

import (
	"flashcards-bot/src/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func processCommand(b *tgBot, update tgbotapi.Update) {
	//Update language code
	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, Language: update.Message.From.LanguageCode}); err != nil {
		b.Logger.Errorw("Error adding language tag", "error", err.Error())
	}

	b.Logger.Infow("Command", "from", update.Message.From.UserName, "body", update.Message.Text)
	switch update.Message.Command() {
	case "start":
		startCommand(b, update)
	case "help":
		helpCommand(b, update)
	case "new_deck":
		newDeckCommand(b, update)
	case "add_cards":
		addCardsCommand(b, update)
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
	_, err := db.GetUser(update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user state", "error", err.Error())
	}

	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Send start message
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.Start[lang])

	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

func helpCommand(b *tgBot, update tgbotapi.Update) {
	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.Help[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

func newDeckCommand(b *tgBot, update tgbotapi.Update) {
	//Update user state to "waiting for a deck name to create new deck"
	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, State: waitingNewDeckName}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Send the next message to the user
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.ChooseDeckName[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}
func addCardsCommand(b *tgBot, update tgbotapi.Update) {
	//Update user state to "waiting for a deck in which card should be created" and put deck name in there
	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, State: waitingNewCardDeckName}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Create a message
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.ChooseDeck[lang])

	//Create an inline keyboard
	keyboard, decksAmount, err := createDecksInlineKeyboard(b, update.Message.From.ID, 1, lang)
	if err != nil {
		b.Logger.Errorw("Error creating a keyboard", "error", err.Error())
	}

	//If user has no decks prompt him to create one first
	if decksAmount <= 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.CreateDeckFirst[lang])
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	// Attaching the keyboard to the message
	msg.ReplyMarkup = keyboard

	// Sending the message with the attached inline keyboard
	sentMessage, err := b.Bot.Send(msg)
	if err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}

	//Add sent message to delete queue to make sure that inline keyboard is deleted later
	b.DeleteQueue = append(b.DeleteQueue, sentMessage.MessageID)
}
func listDecksCommand(b *tgBot, update tgbotapi.Update) {
	//Get decks from db
	decks, err := db.GetDecks(update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting decks from db", "error", err.Error())
	}

	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//If no decks tell user that they have no decks
	if len(decks) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.NoDecks[lang])
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//List user's decks
	var table strings.Builder
	table.WriteString(b.Messages.ListDecks[lang])

	for i := 0; i < len(decks); {
		for j := 0; j < 90 && i < len(decks); j++ {
			table.WriteString(fmt.Sprintf("%d. %v\n", i+1, decks[i]))
			i++
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, table.String())
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		table.Reset()
	}
}

func listCardsCommand(b *tgBot, update tgbotapi.Update) {
	//Update user state to "waiting for a deck name to delete"
	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, State: waitingListMyCardsDeckName}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
		return
	}

	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Create a keyboard with decks
	keyboard, decksAmount, err := createDecksInlineKeyboard(b, update.Message.From.ID, 1, lang)
	if err != nil {
		b.Logger.Errorw("Error creating a keyboard", "error", err.Error())
		return
	}

	//If user has no decks tell them that
	if decksAmount <= 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.NoDecks[lang])
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
	}

	//Prompt user to choose deck
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.ChooseDeck[lang])
	msg.ReplyMarkup = keyboard
	sentMessage, err := b.Bot.Send(msg)
	if err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
	b.DeleteQueue = append(b.DeleteQueue, sentMessage.MessageID)
}

func deleteDeckCommand(b *tgBot, update tgbotapi.Update) {
	//Update user state to "waiting for a deck name to delete"
	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, State: waitingDeleteDeckName}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Create a new message
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.ChooseDeck[lang])

	//Create a new keyboard with decks to choose from
	keyboard, decksAmount, err := createDecksInlineKeyboard(b, update.Message.From.ID, 1, lang)
	if err != nil {
		b.Logger.Errorw("Error getting decks", "error", err.Error())
	}

	//If user has no decks let them now
	if decksAmount == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.NoDecks[lang])
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
			return
		}
		return
	}

	// Sending the message with the attached inline keyboard
	msg.ReplyMarkup = keyboard
	sentMessage, err := b.Bot.Send(msg)
	if err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
		return
	}
	b.DeleteQueue = append(b.DeleteQueue, sentMessage.MessageID)
}

func deleteCardCommand(b *tgBot, update tgbotapi.Update) {
	//Update user state to "waiting for a deck to delete a card from"
	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, State: waitingDeleteCardDeckName}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Create  a keyboard with decks to choose from
	keyboard, decksAmount, err := createDecksInlineKeyboard(b, update.Message.From.ID, 1, lang)
	if err != nil {
		b.Logger.Errorw("Error creating inline keyboard", "error", err.Error())
	}

	//If user has no decks tell them that
	if decksAmount <= 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.NoDecks[lang])
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//Create and send the message
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.ChooseDeck[lang])
	msg.ReplyMarkup = keyboard
	sentMessage, err := b.Bot.Send(msg)
	if err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
		return
	}
	b.DeleteQueue = append(b.DeleteQueue, sentMessage.MessageID)
}

func studyDeckCommand(b *tgBot, update tgbotapi.Update) {
	//Update user state to "waiting cards to study"
	if err := db.UpdateUser(&db.User{TgUserId: update.Message.From.ID, State: waitingStudyDeckName}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Create a keyboard with deck names
	keyboard, decksAmount, err := createDecksInlineKeyboard(b, update.Message.From.ID, 1, lang)
	if err != nil {
		b.Logger.Errorw("Error creating inline keyboard", "error", err.Error())
	}

	//If user has no decks notify user
	if decksAmount <= 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.NoDecks[lang])
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//Add created keyboard to the new message and send it
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.ChooseDeck[lang])
	msg.ReplyMarkup = keyboard
	sentMessage, err := b.Bot.Send(msg)
	if err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
		return
	}
	b.DeleteQueue = append(b.DeleteQueue, sentMessage.MessageID)
}

func unknownCommand(b *tgBot, update tgbotapi.Update) {
	//Get user language
	lang, err := language(update.Message.From.LanguageCode, update.Message.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, b.Messages.UnknownCommand[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}
