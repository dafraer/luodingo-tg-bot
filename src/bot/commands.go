package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func processCommand(b *tgBot, update tgbotapi.Update) {
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
	}
}

func startCommand(b *tgBot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.Start)
	b.bot.Send(msg)
}

func helpCommand(b *tgBot, update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, en.Help)
	b.bot.Send(msg)
}

func newDeckCommand(b *tgBot, update tgbotapi.Update) {

}
func newCardCommand(b *tgBot, update tgbotapi.Update) {

}
func listDecksCommand(b *tgBot, update tgbotapi.Update) {

}

func listCardsCommand(b *tgBot, update tgbotapi.Update) {

}

func deleteDeckCommand(b *tgBot, update tgbotapi.Update) {

}

func deleteCardCommand(b *tgBot, update tgbotapi.Update) {

}

func studyDeckCommand(b *tgBot, update tgbotapi.Update) {

}
