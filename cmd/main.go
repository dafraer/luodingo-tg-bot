package main

import (
	"flashcards-bot/src/bot"
	"flashcards-bot/src/config"
	"flashcards-bot/src/db"
	"fmt"
	"os"
)

func main() {
	config.Load(os.Args[1], os.Args[2])
	if err := db.Connect(config.DatabaseConfig.Host, config.DatabaseConfig.User, config.DatabaseConfig.Password, config.DatabaseConfig.DbName, config.DatabaseConfig.Port); err != nil {
		panic(fmt.Errorf("error connecting to the database: %v", err))
	}

	myBot := bot.New(config.BotConfig.Token, config.BotConfig.Timeout, config.BotConfig.Offset)
	myBot.Run()

	if err := db.Disconnect(); err != nil {
		panic(fmt.Errorf("error disconnecting from db: %v", err))
	}
	if err := myBot.Logger.Sync(); err != nil {
		panic(fmt.Errorf("error flushing buffer: %v", err))
	}

}
