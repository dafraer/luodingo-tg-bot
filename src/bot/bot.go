package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type state uint8

const (
	defaultState   state = iota
	newDeck              //Waiting for deck name to create a new deck
	myCards              //Waiting for deck name to list it's cards
	deleteDeck           //Waiting for deck name to delete deck
	deckDeleteCard       //Waiting for deck name to delete card in that deck
	cardDeleteCard       //Waiting for a card name to delete that card
	deckNewCard          //Waiting for a deck name to create new card in selected deck
	cardNewCard          //Waiting for a  card name to create new card
	studyDeck            //Waiting for a deck name to study
)

type tgBot struct {
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
}

func New(token string, timeout int, offset int) *tgBot {
	myBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(fmt.Errorf("error while creatinga new bot, %v ", err))
	}

	u := tgbotapi.NewUpdate(offset)
	u.Timeout = timeout

	updates := myBot.GetUpdatesChan(u)
	log.Printf("Authorized on account %s", myBot.Self.UserName)
	return &tgBot{
		bot:     myBot,
		updates: updates,
	}
}
