package db

import (
	"fmt"
	"time"
)

type User struct {
	Id           int `gorm:"primaryKey"`
	TgUserId     int64
	StudyTime    time.Duration
	State        int
	DeckSelected string
	CardSelected string
}

type Deck struct {
	Id          int `gorm:"primary_key"`
	TgUserId    int64
	Name        string
	CardsAmount int
}

type Card struct {
	Id      int `gorm:"primaryKey"`
	DeckId  int
	Front   string
	Back    string
	Learned bool //Either learned or not learned
}

func (d Deck) String() string {
	return fmt.Sprintf("%s : %d cards", d.Name, d.CardsAmount)
}
