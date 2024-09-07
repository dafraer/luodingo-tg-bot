package bot

import (
	"flashcards-bot/src/db"
	"flashcards-bot/src/text"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var en, ru text.Messages

func (b *tgBot) Run() {
	en = text.LoadEnMessages()
	ru = text.LoadRuMessages()
	for update := range b.updates {
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

func createDecksInlineKeyboard(b *tgBot, update tgbotapi.Update) (keyboard tgbotapi.InlineKeyboardMarkup, decksAmount int, err error) {
	//Get decks from database
	decks, err := db.GetDecks(update.Message.From.ID)

	//Create buttons
	var buttons [][]tgbotapi.InlineKeyboardButton
	for _, v := range decks {
		buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(v.Name), fmt.Sprint(v.Name))))
	}

	//Return the keyboard with created buttons
	return tgbotapi.NewInlineKeyboardMarkup(buttons...), len(decks), err
}

//*******************************
// REFACTOR EVERYTHING UNDER THIS COMMENT
//*******************************

func newCardHandler(b *tgBot, update tgbotapi.Update) {
	b.sessions.mutex.Lock()
	defer b.sessions.mutex.Unlock()
	var state state
	if update.CallbackQuery != nil {
		state = b.sessions.userStates[update.CallbackQuery.From.ID].action
	} else {
		state = b.sessions.userStates[update.Message.From.ID].action
	}
	switch state {
	case defaultState:
		b.sessions.userStates[update.Message.From.ID] = userState{action: deckNewCard}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseDeck)

		// Creating inline keyboard with buttons
		decks, err := db.GetDecks(update.Message.From.ID)
		if len(decks) == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.CreateDeckFirst)
			b.bot.Send(msg)
			return
		}
		if err != nil {
			log.Printf("ERROR GETTING DECKS:%v\n", err)
		}
		var buttons [][]tgbotapi.InlineKeyboardButton
		for _, v := range decks {
			buttons = append(buttons, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(fmt.Sprint(v.Name), fmt.Sprint(v.Name))))
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons...)
		// Attaching the keyboard to the message
		msg.ReplyMarkup = keyboard

		// Sending the message with the attached inline keyboard
		b.bot.Send(msg)
	case deckNewCard:
		b.sessions.userStates[update.CallbackQuery.From.ID] = userState{action: cardNewCard, deckSelected: update.CallbackQuery.Data}
		edit := tgbotapi.NewEditMessageText(
			update.CallbackQuery.Message.Chat.ID,
			update.CallbackQuery.Message.MessageID,
			en.ChooseCardName,
		)
		b.bot.Send(edit)
	case cardNewCard:
		b.sessions.userStates[update.Message.From.ID] = userState{}
		err := db.CreateCard(b.sessions.userStates[update.Message.From.ID].deckSelected, update.Message.From.ID, update.Message.Text)
		b.sessions.userStates[update.Message.From.ID] = userState{}
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ErrorCreatingCard)
			b.bot.Send(msg)
			return
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.CardCreated)
		b.bot.Send(msg)
	}
}

func listDecksHandler(b *tgBot, update tgbotapi.Update) {
	decks, err := db.GetDecks(update.Message.From.ID)
	if err != nil {
		log.Printf("ERROR GETTING USER DECKS FROM DB: %v", err)
	}

	if len(decks) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.NoDecks)
		b.bot.Send(msg)
		return
	}
	table := "These are your decks:\n"
	for i, v := range decks {
		table += fmt.Sprintf("%d. %v\n", i+1, v)
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, table)
	b.bot.Send(msg)
}

func listCardsHandler(b *tgBot, update tgbotapi.Update) {

}
