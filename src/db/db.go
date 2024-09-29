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
	CardSelected int
	PageSelected int
	Language     string
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

var db *gorm.DB

func Connect(host, user, password, name, port string) (err error) {
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=require sslrootcert=ca.pem", host, user, password, name, port)
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
