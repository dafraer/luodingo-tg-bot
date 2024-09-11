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

func (c Card) String() string {
	return fmt.Sprintf("+--------------------------------------+\n|                                      |\n|  ┌──────────────────────────────┐    |\n|  │                              │    |\n|           %s               |\n|  │                              │    |\n|  └──────────────────────────────┘    |\n|                                      |\n|                                      |\n+--------------------------------------+\n|                                      |\n|  ┌──────────────────────────────┐    |\n|  │                              │    |\n|  │         BACK OF CARD          │    |\n|  │                              │    |\n|  └──────────────────────────────┘    |\n|                                      |\n+--------------------------------------+\n", c.Front, c.Back)
}
