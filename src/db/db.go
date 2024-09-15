package db

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

type Storage interface {
	CreateUser(user *User) error
	GetUser(user *User) (*User, error)
	GetUsers() ([]*User, error)
	UpdateUser(user *User) error
	DeleteUser(user *User) error
	CreateDeck(deck *Deck) error
	GetDeck(deck *Deck) (*Deck, error)
	UpdateDeck(deck *Deck) error
	DeleteDeck(deck *Deck) error
	CreateCard(card *Card) error
	GetCard(card *Card) (*Card, error)
	UpdateCard(card *Card) error
	DeleteCard(card *Card) error
}

var db *gorm.DB

func Connect(host, user, password, name, port string) (err error) {
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=disable", host, user, password, name, port)
	db, err = gorm.Open(postgres.Open(dsn))
	if err != nil {
		return
	}
	err = db.AutoMigrate(&User{}, &Deck{}, &Card{})
	return
}

func Disconnect() (err error) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	// Close the connection
	err = sqlDB.Close()
	return
}

func (d Deck) String() string {
	return fmt.Sprintf("%s : %d cards", d.Name, d.CardsAmount)
}
