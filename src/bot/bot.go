package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

const (
	defaultState               int = iota
	waitingNewDeckName             //Waiting for deck name to create a new deck
	waitingListMyCardsDeckName     //Waiting for deck name to list its cards
	waitingDeleteDeckName          //Waiting for deck name to delete deck
	waitingDeleteCardDeckName      //Waiting for deck name to delete card in that deck
	waitingDeleteCardCardName      //Waiting for a card name to delete that card
	waitingNewCardDeckName         //Waiting for a deck name to create new card in selected deck
	waitingNewCardFront            //Waiting for a  card's front to create new card
	waitingNewCardBack             //Waiting for a card's back to create new card
	waitingStudyDeckName           //Waiting for a deck name to study
	waitingFlipCard                //Waiting for user to flip the card he is studying
	waitingCardFeedback            //Waiting for user to pick if he learned the card or no
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
