package bot

import (
	"flashcards-bot/src/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func processCallback(b *tgBot, update tgbotapi.Update) {
	b.Logger.Infow("Callback query", "from", update.CallbackQuery.From.UserName, "data", update.CallbackQuery.Data)
	user, err := db.GetUserState(update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user state", "error", err.Error())
	}
	switch user.State {
	case waitingDeleteDeckName:
		deleteDeckCallback(b, update)
	case waitingNewCardDeckName:
		newCardCallback(b, update)
	case waitingDeleteCardDeckName:
		deckDeleteCardCallback(b, update)
	case waitingDeleteCardCardName:
		cardDeleteCardCallback(b, update, user.DeckSelected)
	case waitingListMyCardsDeckName:
		listCardsCallback(b, update)
	case waitingStudyDeckName:
		studyDeckCallback(b, update)
	case waitingCardFeedback:
		studyCardCallback(b, update, user)
	default:
		unknownCallback(b, update)
	}
}

// studyDeckCallback is called when user chooses a deck they want to study
// Basically study session begins here
func studyDeckCallback(b *tgBot, update tgbotapi.Update) {

	//Update user state to studying
	if err := db.UpdateUserState(db.User{TgUserId: update.CallbackQuery.From.ID, State: waitingCardFeedback, DeckSelected: update.CallbackQuery.Data}); err != nil {
		b.Logger.Errorw("Error updating deck state", "error", err.Error())
	}

	edit := b.studyRandomCard(update)
	if _, err := b.Bot.Send(edit); err != nil {
		b.Logger.Errorw("Error sending edit", "error", err.Error())
	}
}

func studyCardCallback(b *tgBot, update tgbotapi.Update, user db.User) {
	//If user pressed button to stop studying delete the message to get rid of an inline keyboard
	switch update.CallbackQuery.Data {
	case "stop":
		b.deleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	case "✅":
		if err := db.UpdateCardState(user.CardSelected, user.DeckSelected, update.CallbackQuery.From.ID, true); err != nil {
			b.Logger.Errorw("Error updating card state", "error", err.Error())
		}

		//Go for the next card
		edit := b.studyRandomCard(update)
		if _, err := b.Bot.Send(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}
	case "❎":
		//Go for the next card
		edit := b.studyRandomCard(update)
		if _, err := b.Bot.Send(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}
	default:
		//Default case happens when user asks to show the answer to the card

		//Create buttons first
		checkButton := tgbotapi.NewInlineKeyboardButtonData("✅", "✅")
		crossButton := tgbotapi.NewInlineKeyboardButtonData("❎", "❎")

		//Put both check and cross buttons in the same row
		row := tgbotapi.NewInlineKeyboardRow(checkButton, crossButton)

		//Create "Stop studying button in a separate row"
		stopButton := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(en.StopStudy, "stop"))

		//Create a keyboard using previously created buttons
		keyboard := tgbotapi.NewInlineKeyboardMarkup(row, stopButton)

		//Create and send the edit
		edit := tgbotapi.NewEditMessageTextAndMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			update.CallbackQuery.Message.Text+"\n"+update.CallbackQuery.Data,
			keyboard,
		)
		if _, err := b.Bot.Send(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}

	}
}

func listCardsCallback(b *tgBot, update tgbotapi.Update) {
	//Get cards from db
	cards, err := db.GetCards(update.CallbackQuery.Data, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting cards from db", "error", err.Error())
	}

	//If no cards tell user that they have no cards
	if len(cards) == 0 {
		edit := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			en.NoCards,
		)
		if _, err := b.Bot.Send(edit); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//List user's cards
	table := fmt.Sprintf("These are the cards from %s deck:\n", update.CallbackQuery.Data)
	for i, v := range cards {
		table += fmt.Sprintf("%d. %s-%s\n", i+1, v.Front, v.Back)
	}

	//Made it an edit so inline keyboard disappears
	edit := tgbotapi.NewEditMessageText(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		table,
	)
	if _, err := b.Bot.Send(edit); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

func deleteDeckCallback(b *tgBot, update tgbotapi.Update) {
	//Don't update user state because then callback cause bot to crash

	//Get deck name to delete
	name := update.CallbackQuery.Data

	//Delete the deck
	if err := db.DeleteDeck(name, update.CallbackQuery.From.ID); err != nil {
		b.Logger.Errorw("Error deleting deck", "error", err.Error())

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.ErrorDeletingDeck)
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
	}

	//Get decks keyboard
	keyboard, decksAmount, err := createDecksInlineKeyboard(update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting inline keyboard", "error", err.Error())
	}
	//if user has no decks left delete the message
	if decksAmount <= 0 {
		b.deleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.DeckDeleted)
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//Update inline keyboard
	edit := tgbotapi.NewEditMessageReplyMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		keyboard,
	)
	if _, err := b.Bot.Send(edit); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.DeckDeleted)
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

func unknownCallback(b *tgBot, update tgbotapi.Update) {
	b.Logger.Infow("Unknown callback query", "from", update.CallbackQuery.From.ID, "data", update.CallbackQuery.Data)
}

func newCardCallback(b *tgBot, update tgbotapi.Update) {
	//Update the user state to "waiting card front"
	if err := db.UpdateUserState(db.User{TgUserId: update.CallbackQuery.From.ID, State: waitingNewCardFront, DeckSelected: update.CallbackQuery.Data}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}
	//Edit the message
	edit := tgbotapi.NewEditMessageText(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		en.ChooseCardFront,
	)

	//Send the edit
	if _, err := b.Bot.Send(edit); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

// deckDeleteCardCallback selects a deck in user state and sends an inline keyboard with cards to user
func deckDeleteCardCallback(b *tgBot, update tgbotapi.Update) {
	if err := db.UpdateUserState(db.User{TgUserId: update.CallbackQuery.From.ID, State: waitingDeleteCardCardName, DeckSelected: update.CallbackQuery.Data}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Create inline keyboard of cards in a selected deck
	keyboard, cardsAmount, err := createCardsInlineKeyboard(update.CallbackQuery.From.ID, update.CallbackQuery.Data)
	if err != nil {
		b.Logger.Errorw("Error getting inline keyboard for cards", "error", err.Error())
	}

	//If deck has no cards notify user about it
	if cardsAmount <= 0 {
		edit := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			en.NoCards,
		)
		if _, err := b.Bot.Send(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}
		return
	}

	//Send message with an inline keyboard
	edit := tgbotapi.NewEditMessageReplyMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		keyboard,
	)
	if _, err := b.Bot.Send(edit); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

// cardDeleteCard deletes card
func cardDeleteCardCallback(b *tgBot, update tgbotapi.Update, deckName string) {
	//Delete card
	if err := db.DeleteCard(deckName, update.CallbackQuery.From.ID, update.CallbackQuery.Data); err != nil {
		b.Logger.Error("Error deleting card", err.Error())
		return
	}

	//Create a new inline keyboard without the deleted card
	keyboard, cardsAmount, err := createCardsInlineKeyboard(update.CallbackQuery.From.ID, deckName)
	if err != nil {
		b.Logger.Errorw("Error getting inline keyboard for cards", "error", err.Error())
		return
	}

	//If no cards have left - delete the message with inline keyboard
	if cardsAmount <= 0 {
		b.deleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.CardDeleted)
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//Edit inline keyboard and send the edit
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		en.ChooseCard,
		keyboard,
	)
	if _, err := b.Bot.Send(edit); err != nil {
		b.Logger.Errorw("Error sending edit", "error", err.Error())
		return
	}

	//Notify the user that card has been deleted successfully
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.CardDeleted)
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}
