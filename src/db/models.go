package db

import (
	"gorm.io/gorm"
)

type User struct {
	Id           int `gorm:"primaryKey"`
	TgUserId     int64
	State        int
	DeckSelected string
	CardSelected int
	PageSelected int
	Language     string
}

type Deck struct {
	Id       int `gorm:"primary_key"`
	TgUserId int64
	Name     string
}

type Card struct {
	Id      int `gorm:"primaryKey"`
	DeckId  int
	Front   string
	Back    string
	Learned bool
}

var db *gorm.DB
