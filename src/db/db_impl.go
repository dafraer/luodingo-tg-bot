package db

import (
	"errors"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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

func CreateDeck(name string, userId int64) (err error) {
	result := db.Table("decks").Create(&Deck{TgUserId: userId, Name: name, CardsAmount: 0})
	return result.Error
}

func GetDecks(userId int64) (decks []Deck, err error) {
	result := db.Table("decks").Where("tg_user_id = ?", userId).Find(&decks)
	return decks, result.Error
}

func UpdateDeck(oldDeck, newDeck Deck) (err error) {
	result := db.Table("decks").Where("name = ? and tg_user_id = ?", oldDeck.Name, oldDeck.TgUserId).Updates(newDeck)
	return result.Error
}

func DeleteDeck(name string, userId int64) (err error) {
	//Find deck id
	var deck Deck
	result := db.Table("decks").Select("id").Find(&deck, "name = ? and tg_user_id = ?", name, userId)
	if result.Error != nil {
		return result.Error
	}
	//Delete all cards of that deck
	result = db.Table("cards").Joins("JOIN decks on decks.id = cards.deck_id").Where("deck_id = ?", deck.Id).Delete(&Card{})
	if result.Error != nil {
		return result.Error
	}
	//Delete the deck
	result = db.Table("decks").Where("name = ? and tg_user_id = ?", name, userId).Delete(&Deck{})
	err = result.Error
	return
}

func CreateCard(deckName string, userId int64, front, back string) (err error) {
	//TODO: make this work in 1 request
	var deck Deck
	result := db.Table("decks").Find(&deck, "name = ? and tg_user_id = ?", deckName, userId)
	if result.Error != nil {
		return result.Error
	}
	result = db.Table("cards").Joins("JOIN decks on decks.id = cards.deck_id").Create(&Card{DeckId: deck.Id, Front: front, Back: back, Learned: false})
	if result.Error != nil {
		return result.Error
	}

	//Increase amount of cards in the deck
	deck.CardsAmount++
	db.Table("decks").Where("name = ? and tg_user_id = ?", deckName, userId).Updates(&deck)
	return result.Error
}

func GetCards(deckName string, userId int64) (cards []Card, err error) {
	result := db.Table("cards").Joins("JOIN decks on decks.id = cards.deck_id").Where("decks.name = ? and decks.tg_user_id = ?", deckName, userId).Find(&cards)
	return cards, result.Error
}

func DeleteCard(deckName string, userId int64, cardId string) (err error) {
	//TODO: make this work in 1 request
	var deck Deck
	result := db.Table("decks").Select("id").Find(&deck, "name = ? and tg_user_id = ?", deckName, userId)
	if result.Error != nil {
		return result.Error
	}
	result = db.Table("cards").Joins("JOIN decks on decks.id = cards.deck_id").Where("deck_id = ? and cards.id = ?", deck.Id, cardId).Delete(&Card{})
	return result.Error
}

func GetUserState(userId int64) (user User, err error) {
	result := db.Table("users").Find(&user, "tg_user_id = ?", userId)
	if result.RowsAffected == 0 {
		return User{}, errors.New("user not found")
	}
	return user, result.Error
}

func UpdateUserState(user User) (err error) {
	result := db.Table("users").Where("tg_user_id = ?", user.TgUserId).Updates(user)
	return result.Error
}

func CreateUser(user User) (err error) {
	result := db.Table("users").Create(&user)
	return result.Error
}
