package bot

import (
	"flashcards-bot/src/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

const (
	leftDeck  = "qF5!v6r9Vm"
	rightDeck = "_(dC9z96D#"
	leftCard  = "V4q38!9mZo"
	rightCard = "9r62'Q7]}E"
	stop      = "?9{i6WL6Y|"
	check     = "%4k4!OI0/%"
	cross     = "%MUg0L8<3m"
	done      = "J0z5'3-1GD"
)

func processCallback(b *tgBot, update tgbotapi.Update) {
	b.Logger.Infow("Callback query", "from", update.CallbackQuery.From.UserName, "data", update.CallbackQuery.Data)
	user, err := db.GetUser(update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user state", "error", err.Error())
		return
	}

	//Handle flipping pages in inline keyboard separately because I am dumb fuck who couldnt implement it in a better way
	switch update.CallbackQuery.Data {
	case leftDeck:
		flipDecksCallback(b, update, "left", user)
		return
	case rightDeck:
		flipDecksCallback(b, update, "right", user)
		return
	case leftCard:
		flipCardsCallback(b, update, "left", user)
		return
	case rightCard:
		flipCardsCallback(b, update, "right", user)
		return
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
		studyCardCallback(b, update, *user)
	case waitingNewCardFront:
		doneCallback(b, update)
	default:
		unknownCallback(b, update)
	}
}

// studyDeckCallback is called when user chooses a deck they want to study
// Basically study session begins here
func studyDeckCallback(b *tgBot, update tgbotapi.Update) {

	//Update user state to studying
	if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, State: waitingCardFeedback, DeckSelected: update.CallbackQuery.Data}); err != nil {
		b.Logger.Errorw("Error updating deck state", "error", err.Error())
	}

	edit, err := b.studyRandomCard(update)
	if err != nil {
		b.Logger.Errorw("Error getting edit with random card", "error", err.Error())
		return
	}
	if _, err := b.Bot.Send(edit); err != nil {
		b.Logger.Errorw("Error sending edit", "error", err.Error())
	}
}

func studyCardCallback(b *tgBot, update tgbotapi.Update, user db.User) {
	//If user pressed button to stop studying delete the message to get rid of an inline keyboard
	switch update.CallbackQuery.Data {
	case stop:
		b.deleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	case check:
		if err := db.UpdateCard(user.DeckSelected, update.CallbackQuery.From.ID, user.CardSelected, true); err != nil {
			b.Logger.Errorw("Error updating card state", "error", err.Error())
		}

		//Go for the next card
		edit, err := b.studyRandomCard(update)
		if err != nil {
			b.Logger.Errorw("Error getting edit with random card", "error", err.Error())
			return
		}

		if _, err := b.Bot.Send(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}
	case cross:
		//Go for the next card
		edit, err := b.studyRandomCard(update)
		if err != nil {
			b.Logger.Errorw("Error getting edit with random card", "error", err.Error())
			return
		}

		if _, err := b.Bot.Send(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}
	default:
		//Default case happens when user asks to show the answer to the card

		//Create buttons first
		checkButton := tgbotapi.NewInlineKeyboardButtonData("✅", check)
		crossButton := tgbotapi.NewInlineKeyboardButtonData("❎", cross)

		//Put both check and cross buttons in the same row
		row := tgbotapi.NewInlineKeyboardRow(checkButton, crossButton)

		//Create "Stop studying button in a separate row"
		lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
		if err != nil {
			b.Logger.Errorw("Error getting user language", "error", err.Error())
		}
		stopButton := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.StopStudy[lang], stop))

		//Create a keyboard using previously created buttons
		keyboard := tgbotapi.NewInlineKeyboardMarkup(row, stopButton)

		//Create and send the edit
		edit := tgbotapi.NewEditMessageTextAndMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			fmt.Sprintf("%s\n——————————————————————\n%s", user.CardSelected, update.CallbackQuery.Data),
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

	//Get user language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//If no cards tell user that they have no cards
	if len(cards) == 0 {
		edit := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			b.Messages.NoCards[lang],
		)
		if _, err := b.Bot.Send(edit); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//List user's cards
	var table strings.Builder
	table.WriteString(b.Messages.ListCards[lang])

	for i := 0; i < len(cards); {
		for j := 0; j < 90 && i < len(cards); j++ {
			table.WriteString(fmt.Sprintf("%d. %s-%s\n", i+1, cards[i].Front, cards[i].Back))
			i++
		}
		//Made it an edit so inline keyboard disappears
		b.deleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, table.String())
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		table.Reset()
	}
}

func deleteDeckCallback(b *tgBot, update tgbotapi.Update) {
	//Don't update user state because then callback cause bot to crash
	//Get deck name to delete
	name := update.CallbackQuery.Data

	//Get user language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Delete the deck
	if err := db.DeleteDeck(&db.Deck{Name: name, TgUserId: update.CallbackQuery.From.ID}); err != nil {
		b.Logger.Errorw("Error deleting deck", "error", err.Error())

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, b.Messages.ErrorDeletingDeck[lang])
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
	}

	//Get decks keyboard
	keyboard, decksAmount, err := createDecksInlineKeyboard(update.CallbackQuery.From.ID, 1)
	if err != nil {
		b.Logger.Errorw("Error getting inline keyboard", "error", err.Error())
	}
	//if user has no decks left delete the message
	if decksAmount <= 0 {
		b.deleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, b.Messages.DeckDeleted[lang])
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
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, b.Messages.DeckDeleted[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

func unknownCallback(b *tgBot, update tgbotapi.Update) {
	b.Logger.Infow("Unknown callback query", "from", update.CallbackQuery.From.ID, "data", update.CallbackQuery.Data)
}

func newCardCallback(b *tgBot, update tgbotapi.Update) {
	//Update the user state to "waiting card front"
	if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, State: waitingNewCardFront, DeckSelected: update.CallbackQuery.Data, PageSelected: 1}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Get user language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Create an inline keyboard to stop adding cards
	keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.Done[lang], done)))

	//Edit the message
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		b.Messages.ChooseCardFront[lang],
		keyboard,
	)

	//Send the edit
	if _, err := b.Bot.Request(edit); err != nil {
		b.Logger.Errorw("Error sending edit", "error", err.Error())
	}
}

// deckDeleteCardCallback selects a deck in user state and sends an inline keyboard with cards to user
func deckDeleteCardCallback(b *tgBot, update tgbotapi.Update) {
	if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, State: waitingDeleteCardCardName, DeckSelected: update.CallbackQuery.Data}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Create inline keyboard of cards in a selected deck
	b.Logger.Debugw("Creating cards inline keyboard with cards", "deckName", update.CallbackQuery.Data, "tg_user_id", update.CallbackQuery.From.ID, "page selected", 1)
	keyboard, cardsAmount, err := createCardsInlineKeyboard(update.CallbackQuery.From.ID, update.CallbackQuery.Data, b, 1)
	b.Logger.Debugw("Created inline keyboard with cards", "cards amount", cardsAmount, "error", err)
	if err != nil {
		b.Logger.Errorw("Error getting inline keyboard for cards", "error", err.Error())
	}

	//Get user language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//If deck has no cards notify user about it
	if cardsAmount <= 0 {
		edit := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			b.Messages.NoCards[lang],
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
	sentMessage, err := b.Bot.Send(edit)
	if err != nil {
		b.Logger.Errorw("Error sending edit", "error", err.Error())
		return
	}
	b.DeleteQueue = append(b.DeleteQueue, sentMessage.MessageID)
}

// cardDeleteCard deletes card
func cardDeleteCardCallback(b *tgBot, update tgbotapi.Update, deckName string) {
	//Delete card
	b.Logger.Debugw("Deleting card", "deck name", deckName, "card id", update.CallbackQuery.Data)
	user, err := db.GetUser(update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user", "error", err.Error())
		return
	}
	if err := db.DeleteCard(update.CallbackQuery.Data, user.DeckSelected, user.TgUserId); err != nil {
		b.Logger.Error("Error deleting card", err.Error())
		return
	}

	//Create a new inline keyboard without the deleted card
	b.Logger.Debugw("Creating cards inline keyboard with cards", "deckName", deckName, "tg_user_id", update.CallbackQuery.From.ID)
	keyboard, cardsAmount, err := createCardsInlineKeyboard(update.CallbackQuery.From.ID, deckName, b, user.PageSelected)
	if err != nil {
		b.Logger.Errorw("Error getting inline keyboard for cards", "error", err.Error())
		return
	}

	//Get user language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//If no cards have left - delete the message with inline keyboard
	if cardsAmount <= 0 {
		b.deleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)

		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, b.Messages.CardDeleted[lang])
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//Edit inline keyboard and send the edit
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		b.Messages.ChooseCard[lang],
		keyboard,
	)
	if _, err := b.Bot.Send(edit); err != nil {
		b.Logger.Errorw("Error sending edit", "error", err.Error())
		return
	}

	//Notify the user that card has been deleted successfully
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, b.Messages.CardDeleted[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

// doneCallback stops the process of adding new cards
func doneCallback(b *tgBot, update tgbotapi.Update) {
	//Get user language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Notify the user about creating card
	msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, b.Messages.CardCreated[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}

	b.deleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
	if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, State: defaultState}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}
}

func flipCardsCallback(b *tgBot, update tgbotapi.Update, direction string, user *db.User) {
	b.Logger.Debugw("flipCardsCallback function called", "page selected", user.PageSelected, "direction", direction)
	if direction == "right" {
		if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, PageSelected: user.PageSelected + 1}); err != nil {
			b.Logger.Errorw("Error updating user state", "error", err.Error())
		}
		keyboard, _, err := createCardsInlineKeyboard(update.CallbackQuery.From.ID, user.DeckSelected, b, user.PageSelected+1)
		if err != nil {
			b.Logger.Errorw("Error getting inline keyboard for cards", "error", err.Error())
		}
		edit := tgbotapi.NewEditMessageReplyMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			keyboard,
		)
		if _, err := b.Bot.Request(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}
	} else {
		if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, PageSelected: user.PageSelected - 1}); err != nil {
			b.Logger.Errorw("Error updating user state", "error", err.Error())
		}
		keyboard, _, err := createCardsInlineKeyboard(update.CallbackQuery.From.ID, user.DeckSelected, b, user.PageSelected-1)
		if err != nil {
			b.Logger.Errorw("Error getting inline keyboard for cards", "error", err.Error())
		}
		edit := tgbotapi.NewEditMessageReplyMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			keyboard,
		)
		if _, err := b.Bot.Request(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}
	}
}

func flipDecksCallback(b *tgBot, update tgbotapi.Update, direction string, user *db.User) {
	if direction == "right" {
		if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, PageSelected: user.PageSelected + 1}); err != nil {
			b.Logger.Errorw("Error updating user state", "error", err.Error())
		}
		keyboard, _, err := createDecksInlineKeyboard(update.CallbackQuery.From.ID, user.PageSelected+1)
		if err != nil {
			b.Logger.Errorw("Error getting inline keyboard for decks", "error", err.Error())
		}
		edit := tgbotapi.NewEditMessageReplyMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			keyboard,
		)
		if _, err := b.Bot.Request(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}
	} else {
		if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, PageSelected: user.PageSelected - 1}); err != nil {
			b.Logger.Errorw("Error updating user state", "error", err.Error())
		}
		keyboard, _, err := createDecksInlineKeyboard(update.CallbackQuery.From.ID, user.PageSelected-1)
		if err != nil {
			b.Logger.Errorw("Error getting inline keyboard for decks", "error", err.Error())
		}
		edit := tgbotapi.NewEditMessageReplyMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			keyboard,
		)
		if _, err := b.Bot.Request(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}
	}
}
