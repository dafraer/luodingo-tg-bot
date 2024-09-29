package config

import (
	_ "github.com/joho/godotenv/autoload"
	"os"
)

type Bot struct {
	Token string
}

type Database struct {
	Host     string
	User     string
	Password string
	DbName   string
	Port     string
}

var BotConfig Bot
var DatabaseConfig Database

func Load() {
	BotConfig = loadBotConfig()
	DatabaseConfig = loadDatabaseConfig()
}

func loadBotConfig() Bot {
	return Bot{
		Token: os.Getenv("BOT_TOKEN"),
	}
}

func loadDatabaseConfig() Database {
	return Database{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		DbName:   os.Getenv("DB_NAME"),
		User:     os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
	}
}
