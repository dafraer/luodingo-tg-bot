package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Bot struct {
	Token   string
	Offset  int
	Timeout int
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

func Load(botPath, dbPath string) {
	BotConfig = loadBotConfig(botPath)
	DatabaseConfig = loadDatabaseConfig(dbPath)
}

func loadBotConfig(path string) Bot {
	var cfg Bot
	config, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("error opening config file: %v", err))
	}
	defer func() {
		if err := config.Close(); err != nil {
			panic(fmt.Errorf("error closing config fil:, %v", err))
		}
	}()

	p, err := io.ReadAll(config)
	if err != nil {
		panic(fmt.Errorf("error opening config file: %v", err))
	}
	if err = json.Unmarshal(p, &cfg); err != nil {
		panic(fmt.Errorf("error parsing config file: %v", err))
	}
	return cfg
}

func loadDatabaseConfig(path string) Database {
	var cfg Database
	config, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("error opening config file: %v", err))
	}
	defer func() {
		if err := config.Close(); err != nil {
			panic(fmt.Errorf("error closing config file: %v", err))
		}
	}()

	p, err := io.ReadAll(config)
	if err != nil {
		panic(fmt.Errorf("error when opening config file, %v", err))
	}
	if err = json.Unmarshal(p, &cfg); err != nil {
		panic(fmt.Errorf("error parsing config file: %v", err))
	}
	return cfg
}
