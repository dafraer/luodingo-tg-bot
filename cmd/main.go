package main

import (
	"flashcards-bot/src/bot"
	"flashcards-bot/src/config"
	"flashcards-bot/src/db"
	"fmt"
)

func main() {
	cfg := config.Load()
	if err := db.Connect(cfg.DbUri); err != nil {
		panic(fmt.Errorf("error connecting to the database: %v", err))
	}

	myBot := bot.New(cfg.BotToken, 60, 0)
	myBot.Logger.Infow("Authorised", "Account", myBot.Bot.Self.UserName)
	myBot.Run()

	if err := db.Disconnect(); err != nil {
		panic(fmt.Errorf("error disconnecting from db: %v", err))
	}
	//flush logger buffer
	if err := myBot.Logger.Sync(); err != nil {
		panic(fmt.Errorf("error flushing buffer: %v", err))
	}
}
