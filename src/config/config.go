package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	BotToken string
	DbUri    string
}

func Load() *Config {
	return &Config{
		BotToken: os.Getenv("BOT_TOKEN"),
		DbUri:    os.Getenv("DB_URI"),
	}
}
