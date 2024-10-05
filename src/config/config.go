package config

import (
	_ "github.com/joho/godotenv/autoload"
	"os"
)

type Config struct {
	BotToken   string
	DbHost     string
	DbUser     string
	DbPassword string
	DbName     string
	DbPort     string
}

func Load() *Config {
	return &Config{
		BotToken:   os.Getenv("BOT_TOKEN"),
		DbHost:     os.Getenv("DB_HOST"),
		DbPort:     os.Getenv("DB_PORT"),
		DbName:     os.Getenv("DB_NAME"),
		DbUser:     os.Getenv("DB_USERNAME"),
		DbPassword: os.Getenv("DB_PASSWORD"),
	}
}
