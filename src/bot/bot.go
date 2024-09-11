package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
	"log"
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
	waitingFlipCard            //Waiting for user to flip the card he is studying
	waitingCardFeedback        //Waiting for user to pick if he learned the card or no
)

type tgBot struct {
	Bot     *tgbotapi.BotAPI
	Updates tgbotapi.UpdatesChannel
	Logger  *zap.SugaredLogger
}

func New(token string, timeout int, offset int) *tgBot {
	myBot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(fmt.Errorf("error while creatinga new Bot, %v ", err))
	}

	u := tgbotapi.NewUpdate(offset)
	u.Timeout = timeout
	updates := myBot.GetUpdatesChan(u)

	logger, err := zap.NewDevelopment()
	sugar := logger.Sugar()

	if err != nil {
		panic(fmt.Errorf("error while creating new Logger, %v ", err))
	}
	log.Printf("Authorized on account %s", myBot.Self.UserName)
	return &tgBot{
		Bot:     myBot,
		Updates: updates,
		Logger:  sugar,
	}
}
