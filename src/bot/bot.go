package bot

import (
	"flashcards-bot/src/db"
	"flashcards-bot/src/text"
	"fmt"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

const (
	_ int = iota
	defaultState
	waitingNewDeckName         //Waiting for deck name to create a new deck
	waitingListMyCardsDeckName //Waiting for deck name to list its cards
	waitingDeleteDeckName      //Waiting for deck name to delete deck
	waitingDeleteCardDeckName  //Waiting for deck name to delete card in that deck
	waitingDeleteCardCardName  //Waiting for a card name to delete that card
	waitingNewCardDeckName     //Waiting for a deck name to create new card in selected deck
	waitingNewCardFront        //Waiting for a  card's front to create new card
	waitingNewCardBack         //Waiting for a card's back to create new card
	waitingStudyDeckName       //Waiting for a deck name to study
	waitingCardFeedback        //Waiting for user to pick if he learned the card or no

	leftDeck           = "qF5!v6r9Vm"
	rightDeck          = "_(dC9z96D#"
	leftCard           = "V4q38!9mZo"
	rightCard          = "9r62'Q7]}E"
	stop               = "?9{i6WL6Y|"
	check              = "%4k4!OI0/%"
	cross              = "%MUg0L8<3m"
	done               = "J0z5'3-1GD"
	cancel             = "H9fj7'10d"
	addReverse         = "gjDHfjFn)dKj"
	maxLinesPerMessage = 90
	maxMessageLen      = 40 //characters
	maxMessageSize     = 64 //bytes
)

type tgBot struct {
	Bot         *tgbotapi.BotAPI
	Updates     tgbotapi.UpdatesChannel
	DeleteQueue []message //Queue to delete messages with inline keyboards
	Logger      *zap.SugaredLogger
	Messages    *text.Messages
}

type message struct {
	msgId  int
	chatId int64
}

func New(token string, timeout int, offset int) *tgBot {
	//Create the bot
	myBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(fmt.Errorf("error while creating a new Bot, %v ", err))
	}

	//Create updates channel
	u := tgbotapi.NewUpdate(offset)
	u.Timeout = timeout
	updates := myBot.GetUpdatesChan(u)

	//Create logger
	logger, err := zap.NewDevelopment()
	var sugar *zap.SugaredLogger
	if logger != nil {
		sugar = logger.Sugar()
	}

	if err != nil {
		panic(fmt.Errorf("error while creating new Logger, %v ", err))
	}

	//Create messages
	messages := text.Load()

	return &tgBot{
		Bot:      myBot,
		Updates:  updates,
		Logger:   sugar,
		Messages: messages,
	}
}

func (b *tgBot) Run() {
	for update := range b.Updates {
		handleUpdates(b, update)
	}
}

func handleUpdates(b *tgBot, update tgbotapi.Update) {
	switch {
	case update.Message != nil && update.Message.IsCommand():
		processCommand(b, update)
	case update.Message != nil:
		processMessage(b, update)
	case update.CallbackQuery != nil:
		processCallback(b, update)
	}
}

// decksInlineKeyboard creates new inline keyboard with decks
func decksInlineKeyboard(b *tgBot, userId int64, page int, lang string) (keyboard tgbotapi.InlineKeyboardMarkup, decksAmount int, err error) {
	//Get decks from database
	decks, err := db.GetDecks(userId)

	//Figure out from which card to display
	from := (page - 1) * 10

	//Create buttons
	var buttons [][]tgbotapi.InlineKeyboardButton
	for i := from; i < min(len(decks), from+10); i++ {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(decks[i].Name), fmt.Sprint(decks[i].Name))))
	}

	//add change page button
	if len(decks) >= 11 {
		switch {
		case from == 0:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➡️️", rightDeck)))
		case from >= len(decks)-10:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", leftDeck)))
		default:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", leftDeck), tgbotapi.NewInlineKeyboardButtonData("➡️️", rightDeck)))
		}
	}
	//Add cancel button
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.Cancel[lang], cancel)))

	//Return the keyboard with created buttons
	return tgbotapi.NewInlineKeyboardMarkup(buttons...), len(decks), err
}

func cardsInlineKeyboard(userId int64, deckName string, b *tgBot, page int, lang string) (keyboard tgbotapi.InlineKeyboardMarkup, cardsAmount int, err error) {
	b.Logger.Debugw("cardsInlineKeyboard", "page", page)
	//Get cards from database
	cards, err := db.GetCards(deckName, userId)

	//Figure out from which card to display
	from := (page - 1) * 10

	//Create buttons with front-back of a card shown to the user and card id sent as a callback data
	var buttons [][]tgbotapi.InlineKeyboardButton
	for i := from; i < min(from+10, len(cards)); i++ {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s-%s", cards[i].Front, cards[i].Back), fmt.Sprint(cards[i].Id))))
	}
	//add change page button
	if len(cards) >= 11 {
		switch {
		case from == 0:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➡️️", rightCard)))
		case from >= len(cards)-10:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", leftCard)))
		default:
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("⬅️", leftCard), tgbotapi.NewInlineKeyboardButtonData("➡️️", rightCard)))
		}
	}
	//Add cancel button
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.Cancel[lang], cancel)))

	//Return the keyboard with created buttons
	return tgbotapi.NewInlineKeyboardMarkup(buttons...), len(cards), err
}

func (b *tgBot) clearDeleteQueue(chatId int64) {
	for _, msg := range b.DeleteQueue {
		if chatId == msg.chatId {
			deleteMessage := tgbotapi.NewDeleteMessage(msg.chatId, msg.msgId)
			if _, err := b.Bot.Request(deleteMessage); err != nil {
				b.Logger.Errorw("Error deleting message", "error", err.Error())
			}
		}
	}
	b.DeleteQueue = nil
}

// studyRandomCard Creates a message with a card to study
func studyRandomCard(b *tgBot, update tgbotapi.Update) (tgbotapi.EditMessageTextConfig, error) {
	//Get user state to know what to know selected deck
	user, err := db.GetUser(update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user state", "error", err.Error())
		return tgbotapi.EditMessageTextConfig{}, err
	}

	lang, err := language(update.CallbackQuery.From.LanguageCode, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting user language", "error", err.Error())
	}

	//Get cards from the selected deck
	cards, err := db.GetUnlearnedCards(user.DeckSelected, update.CallbackQuery.From.ID)
	if err != nil {
		b.Logger.Errorw("Error getting cards", "error", err.Error())
	}

	//If not enough cards tell the user
	if len(cards) == 0 {
		if err := db.UnlearnCards(user.DeckSelected, update.CallbackQuery.From.ID); err != nil {
			b.Logger.Errorw("Error unlearning cards", "error", err.Error())
		}

		edit := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			b.Messages.FinishedStudy[lang],
		)
		return edit, nil
	}

	//Pick a random card
	card := cards[rand.Intn(len(cards))]

	//Create buttons with 2 options:
	//Show back of the card
	//Stop studying
	var buttons [][]tgbotapi.InlineKeyboardButton
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.ShowAnswer[lang], fmt.Sprint(card.Back))))
	buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(b.Messages.StopStudy[lang], stop)))

	//Created an inline keyboard with previously created buttons
	keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)

	//Edit already existing message to remove previous inline keyboard
	edit := tgbotapi.NewEditMessageTextAndMarkup(
		update.CallbackQuery.Message.Chat.ID,
		update.CallbackQuery.Message.MessageID,
		card.Front+"\n——————————————————————",
		keyboard,
	)

	user.CardSelected = card.Id
	if err := db.UpdateUser(user); err != nil {
		b.Logger.Errorw("Error updating user state", "error", err.Error())
	}
	return edit, nil
}

// language figures out user's language. It prioritises language code provided by telegram
func language(languageCode string, userId int64) (string, error) {
	if languageCode != "en" && languageCode != "ru" && languageCode != "es" {
		user, err := db.GetUser(userId)
		if err != nil {
			return "", err
		}

		if user.Language != "en" && user.Language != "ru" && user.Language != "es" {
			return "en", nil
		}
		return user.Language, nil
	}
	return languageCode, nil
}
