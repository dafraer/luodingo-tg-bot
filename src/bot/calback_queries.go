package bot

import (
	"flashcards-bot/src/db"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
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
		flipDecksCallback(b, update, -1, user)
		return
	case rightDeck:
		flipDecksCallback(b, update, 1, user)
		return
	case leftCard:
		flipCardsCallback(b, update, -1, user)
		return
	case rightCard:
		flipCardsCallback(b, update, 1, user)
		return
	case cancel:
		cancelCallback(b, update)
		return
	case addReverse:
		addReverseCardCallback(b, update, user)
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
		cardDeleteCardCallback(b, update, user)
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

func cancelCallback(b *tgBot, update tgbotapi.Update) {
	if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, PageSelected: 1}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}
	b.DeleteQueue = append(b.DeleteQueue, message{update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID})
	b.clearDeleteQueue()
}

func addReverseCardCallback(b *tgBot, update tgbotapi.Update, user *db.User) {
	if user != nil {
		card, err := db.GetCard(user.CardSelected)
		if err != nil {
			b.Logger.Errorw("Error getting card", "error", err.Error())
			return
		}
		if _, err := db.CreateCard(user.DeckSelected, user.TgUserId, &db.Card{Front: card.Back, Back: card.Front, Learned: false}); err != nil {
			b.Logger.Errorw("Error creating card", "error", err.Error())
		}
	}

	//Get user language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
		return
	}

	//Create an edited inline keyboard without function to add reverse card
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.Done[lang], done)),
	)

	edit := tgbotapi.NewEditMessageReplyMarkup(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, keyboard)
	if _, err := b.Bot.Request(edit); err != nil {
		b.Logger.Errorw("Error editing message", "error", err.Error())
		return
	}

	//Notify user that reverse card has been added
	msg := tgbotapi.NewMessage(update.CallbackQuery.From.ID, b.Messages.ReverseAdded[lang])
	if _, err := b.Bot.Send(msg); err != nil {
		b.Logger.Errorw("Error sending message", "error", err.Error())
	}
}

// studyDeckCallback is called when user chooses a deck they want to study
// Basically study session begins here
func studyDeckCallback(b *tgBot, update tgbotapi.Update) {
	//Update user state to studying
	if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, State: waitingCardFeedback, DeckSelected: update.CallbackQuery.Data}); err != nil {
		b.Logger.Errorw("Error updating deck state", "error", err.Error())
	}

	edit, err := studyRandomCard(b, update)
	if err != nil {
		b.Logger.Errorw("Error getting edit with random card", "error", err.Error())
		return
	}
	if _, err := b.Bot.Send(edit); err != nil {
		b.Logger.Errorw("Error sending edit", "error", err.Error())
	}
}

func studyCardCallback(b *tgBot, update tgbotapi.Update, user db.User) {
	switch update.CallbackQuery.Data {
	case stop:
		b.DeleteQueue = append(b.DeleteQueue, message{update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID})
		b.clearDeleteQueue()
	case check:
		if err := db.UpdateCard(&db.Card{Id: user.CardSelected, Learned: true}); err != nil {
			b.Logger.Errorw("Error updating card state", "error", err.Error())
		}

		//Go for the next card
		edit, err := studyRandomCard(b, update)
		if err != nil {
			b.Logger.Errorw("Error getting edit with random card", "error", err.Error())
			return
		}

		if _, err := b.Bot.Request(edit); err != nil {
			b.Logger.Errorw("Error sending edit", "error", err.Error())
		}
	case cross:
		//Go for the next card
		edit, err := studyRandomCard(b, update)
		if err != nil {
			b.Logger.Errorw("Error getting edit with random card", "error", err.Error())
			return
		}

		if _, err := b.Bot.Request(edit); err != nil {
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
		card, err := db.GetCard(user.CardSelected)
		if err != nil {
			b.Logger.Errorw("Error getting card", "error", err.Error())
			return
		}
		//Create and send the edit
		edit := tgbotapi.NewEditMessageTextAndMarkup(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			fmt.Sprintf("%s\n——————————————————————\n%s", card.Front, update.CallbackQuery.Data),
			keyboard,
		)
		if _, err := b.Bot.Request(edit); err != nil {
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
		if _, err := b.Bot.Request(edit); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		return
	}

	//List user's cards
	var list strings.Builder
	list.WriteString(b.Messages.ListCards[lang])

	//Embedded loop is needed to display cards in different messages if we have more than 90 cards
	for i := 0; i < len(cards); {
		for j := 0; j < maxLinesPerMessage && i < len(cards); j++ {
			list.WriteString(strconv.Itoa(i + 1))
			list.WriteString(". ")
			list.WriteString(cards[i].Front)
			list.WriteString("-")
			list.WriteString(cards[i].Back)
			list.WriteString("\n")
			i++
		}
		//Delete message with inline keyboard
		b.DeleteQueue = append(b.DeleteQueue, message{update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID})
		b.clearDeleteQueue()

		//Create and send a message with cards list
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, list.String())
		if _, err := b.Bot.Send(msg); err != nil {
			b.Logger.Errorw("Error sending message", "error", err.Error())
		}
		list.Reset()
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
	keyboard, decksAmount, err := decksInlineKeyboard(b, update.CallbackQuery.From.ID, 1, lang)
	if err != nil {
		b.Logger.Errorw("Error getting inline keyboard", "error", err.Error())
	}
	//if user has no decks left delete the message
	if decksAmount <= 0 {
		b.DeleteQueue = append(b.DeleteQueue, message{update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID})
		b.clearDeleteQueue()

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
	if _, err := b.Bot.Request(edit); err != nil {
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

// deckDeleteCardCallback sends an inline keyboard with cards to user
func deckDeleteCardCallback(b *tgBot, update tgbotapi.Update) {
	if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, State: waitingDeleteCardCardName, DeckSelected: update.CallbackQuery.Data}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}

	//Get language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Create inline keyboard of cards in a selected deck
	keyboard, cardsAmount, err := cardsInlineKeyboard(update.CallbackQuery.From.ID, update.CallbackQuery.Data, b, 1, lang)
	if err != nil {
		b.Logger.Errorw("Error getting inline keyboard for cards", "error", err.Error())
	}

	//If deck has no cards notify user about it
	if cardsAmount <= 0 {
		edit := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			b.Messages.NoCards[lang],
		)
		if _, err := b.Bot.Request(edit); err != nil {
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
	b.DeleteQueue = append(b.DeleteQueue, message{sentMessage.MessageID, update.CallbackQuery.Message.Chat.ID})
	b.clearDeleteQueue()
}

// cardDeleteCard deletes card
func cardDeleteCardCallback(b *tgBot, update tgbotapi.Update, user *db.User) {
	//Delete card
	if err := db.DeleteCard(update.CallbackQuery.Data, user.DeckSelected, user.TgUserId); err != nil {
		b.Logger.Error("Error deleting card", err.Error())
		return
	}

	//Get user language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Create a new inline keyboard without the deleted card
	keyboard, cardsAmount, err := cardsInlineKeyboard(update.CallbackQuery.From.ID, user.DeckSelected, b, user.PageSelected, lang)
	if err != nil {
		b.Logger.Errorw("Error getting inline keyboard for cards", "error", err.Error())
		return
	}

	//If no cards have left - delete the message with inline keyboard
	if cardsAmount <= 0 {
		b.DeleteQueue = append(b.DeleteQueue, message{update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID})
		b.clearDeleteQueue()
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
	if _, err := b.Bot.Request(edit); err != nil {
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
	b.DeleteQueue = append(b.DeleteQueue, message{update.CallbackQuery.Message.MessageID, update.CallbackQuery.Message.Chat.ID})
	b.clearDeleteQueue()
	if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, State: defaultState}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}
}

func flipCardsCallback(b *tgBot, update tgbotapi.Update, direction int, user *db.User) {
	//Get user language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}
	if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, PageSelected: user.PageSelected + direction}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}
	keyboard, _, err := cardsInlineKeyboard(update.CallbackQuery.From.ID, user.DeckSelected, b, user.PageSelected+direction, lang)
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

func flipDecksCallback(b *tgBot, update tgbotapi.Update, direction int, user *db.User) {
	//Get user language
	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	if err := db.UpdateUser(&db.User{TgUserId: update.CallbackQuery.From.ID, PageSelected: user.PageSelected + direction}); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}
	keyboard, _, err := decksInlineKeyboard(b, update.CallbackQuery.From.ID, user.PageSelected+direction, lang)
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
