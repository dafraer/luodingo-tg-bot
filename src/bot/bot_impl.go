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

// sdfdsf
// sdfsdf
// sdfsdf

func unknownMessageHandler(b *tgBot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.UnknownMessage)
	b.bot.Send(msg)

}

func newDeckHandler(b *tgBot, update tgbotapi.Update) {
	//Two conditions: if user just pressed the command, we prompt them to type new deck's name, else we add the deck with the provided name
	b.sessions.mutex.Lock()
	defer b.sessions.mutex.Unlock()

	if b.sessions.userStates[update.Message.From.ID].action == newDeck {
		db.CreateDeck(update.Message.Text, update.Message.From.ID)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.DeckCreated)
		b.bot.Send(msg)
		b.sessions.userStates[update.Message.From.ID] = userState{}
	} else {
		b.sessions.userStates[update.Message.From.ID] = userState{newDeck, "", ""}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseDeckName)
		b.bot.Send(msg)
	}
}

func deleteDeckHandler(b *tgBot, update tgbotapi.Update) {
	b.sessions.mutex.Lock()
	defer b.sessions.mutex.Unlock()

	if update.CallbackQuery != nil {
		name := update.CallbackQuery.Data
		if err := db.DeleteDeck(name, update.CallbackQuery.From.ID); err != nil {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.ErrorDeleteingDeck)
			b.bot.Send(msg)
		} else {
			decks, err := db.GetDecks(update.CallbackQuery.From.ID)
			if len(decks) == 0 {
				delete := tgbotapi.NewDeleteMessage(
					update.CallbackQuery.Message.Chat.ID,
					update.CallbackQuery.Message.MessageID,
				)
				b.bot.Send(delete)
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.DeckDeleted)
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
			edit := tgbotapi.NewEditMessageReplyMarkup(
				update.CallbackQuery.Message.Chat.ID,
				update.CallbackQuery.Message.MessageID,
				keyboard,
			)
			b.bot.Send(edit)
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, en.DeckDeleted)
			b.bot.Send(msg)
			b.sessions.userStates[update.CallbackQuery.From.ID] = userState{}
		}
	} else {
		b.sessions.userStates[update.Message.From.ID] = userState{action: deleteDeck}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.ChooseDeck)

		// Creating inline keyboard with buttons
		decks, err := db.GetDecks(update.Message.From.ID)
		if len(decks) == 0 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.NoDecks)
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
	}
}

func studyDeckHandler(b *tgBot, update tgbotapi.Update) {

}

func deleteCardHandler(b *tgBot, update tgbotapi.Update) {

}

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
