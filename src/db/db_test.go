package db

import (
	"flashcards-bot/src/config"
	"testing"
)

func TestConnect(t *testing.T) {
	config.Load("../config/configs/bot.json", "../config/configs/db.json")
	if err := Connect(config.DatabaseConfig.Host, config.DatabaseConfig.User, config.DatabaseConfig.Password, config.DatabaseConfig.DbName, config.DatabaseConfig.Port); err != nil {
		t.Fatalf("unable to connect to the database: %v", err)
	}
	defer func() {
		err := Disconnect()
		if err != nil {
			t.Fatalf("error disconnecting from db: %v", err)
		}
	}()
}

func TestDeck(t *testing.T) {
	clearDB(t)

	config.Load("../config/configs/bot.json", "../config/configs/db.json")
	if err := Connect(config.DatabaseConfig.Host, config.DatabaseConfig.User, config.DatabaseConfig.Password, config.DatabaseConfig.DbName, config.DatabaseConfig.Port); err != nil {
		t.Fatalf("error connecting to the database: %v", err)
	}
	defer func() {
		err := Disconnect()
		if err != nil {
			t.Fatalf("error disconnecting from db: %v", err)
		}
	}()

	if err := CreateDeck("spanish", 1); err != nil {
		t.Fatalf("unable to create deck: %v", err)
	}
	if err := UpdateDeck(Deck{Name: "spanish", TgUserId: 1}, Deck{Name: "english", TgUserId: 1}); err != nil {
		t.Fatalf("unable to update deck: %v", err)
	}
	deck, err := GetDecks(1)
	if err != nil {
		t.Fatalf("unable to get decks: %v", err)
	}
	if len(deck) != 1 {
		t.Fatalf("expected 1 deck, got %d", len(deck))
	}
	if deck[0].TgUserId != 1 {
		t.Fatalf("deck does not contain tg_user_id 1, got %d", deck[0].TgUserId)
	}
	if deck[0].Name != "english" {
		t.Fatalf("deck does not contain name english, got %s", deck[0].Name)
	}
	if err := DeleteDeck("english", 1); err != nil {
		t.Fatalf("unable to delete deck: %v", err)
	}
	deck, err = GetDecks(1)
	if err != nil {
		t.Fatalf("unable to get decks: %v", err)
	}
	if len(deck) != 0 {
		t.Fatalf("expected 0 decks, got %d", len(deck))
	}
}

func TestCard(t *testing.T) {
	clearDB(t)

	config.Load("../config/configs/bot.json", "../config/configs/db.json")
	if err := Connect(config.DatabaseConfig.Host, config.DatabaseConfig.User, config.DatabaseConfig.Password, config.DatabaseConfig.DbName, config.DatabaseConfig.Port); err != nil {
		t.Fatalf("error connecting to the database: %v", err)
	}
	defer func() {
		err := Disconnect()
		if err != nil {
			t.Fatalf("error disconnecting from db: %v", err)
		}
	}()

	if err := CreateDeck("spanish", 1); err != nil {
		t.Fatalf("unable to create deck: %v", err)
	}
	if err := CreateCard("spanish", 1, "hola", "hello"); err != nil {
		t.Fatalf("unable to create card: %v", err)
	}
	cards, err := GetCards("spanish", 1)
	if err != nil {
		t.Fatalf("unable to get cards: %v", err)
	}
	if len(cards) != 1 {
		t.Fatalf("expected 1 card, got %d", len(cards))
	}
	if cards[0].Front != "hola" {
		t.Fatalf("card does not contain hola, got %s", cards[0].Front)
	}
	if cards[0].Back != "hello" {
		t.Fatalf("card does not contain hello, got %s", cards[0].Back)
	}
	if err := DeleteCard("spanish", 1, "1"); err != nil {
		t.Fatalf("unable to delete card: %v", err)
	}
	cards, err = GetCards("spanish", 1)
	if err != nil {
		t.Fatalf("unable to get cards: %v", err)
	}
	if len(cards) != 0 {
		t.Fatalf("expected 0 cards, got %d", len(cards))
	}
}

func TestUser(t *testing.T) {
	clearDB(t)

	config.Load("../config/configs/bot.json", "../config/configs/db.json")
	if err := Connect(config.DatabaseConfig.Host, config.DatabaseConfig.User, config.DatabaseConfig.Password, config.DatabaseConfig.DbName, config.DatabaseConfig.Port); err != nil {
		t.Fatalf("unable to connect to the database: %v", err)
	}
	defer func() {
		err := Disconnect()
		if err != nil {
			t.Fatalf("error disconnecting from db: %v", err)
		}
	}()
	if err := CreateUser(User{Id: 1, TgUserId: 1, State: 0}); err != nil {
		t.Fatalf("error creating new user: %v", err)
	}

	user, err := GetUserState(1)
	if err != nil {
		t.Fatalf("error getting user state: %v", err)
	}
	if user.Id != 1 {
		t.Fatalf("user id is not 1, got %v", user.TgUserId)
	}
	if user.TgUserId != 1 {
		t.Fatalf("tg_user_id is not 1, got %v", user.TgUserId)
	}
	if user.State != 0 {
		t.Fatalf("user state is not 0, got %v", user.State)
	}

	if err := UpdateUserState(User{Id: 1, TgUserId: 1, State: 2}); err != nil {
		t.Fatalf("error updating user state: %v", err)
	}

	user, err = GetUserState(1)
	if err != nil {
		t.Fatalf("error getting user state: %v", err)
	}
	if user.Id != 1 {
		t.Fatalf("user id is not 1, got %v", user.TgUserId)
	}
	if user.TgUserId != 1 {
		t.Fatalf("tg_user_id is not 1, got %v", user.TgUserId)
	}
	if user.State != 2 {
		t.Fatalf("user state is not 2, got %v", user.State)
	}
}

func clearDB(t *testing.T) {
	config.Load("../config/configs/bot.json", "../config/configs/db.json")
	if err := Connect(config.DatabaseConfig.Host, config.DatabaseConfig.User, config.DatabaseConfig.Password, config.DatabaseConfig.DbName, config.DatabaseConfig.Port); err != nil {
		t.Fatalf("error connecting to the database: %v", err)
	}
	defer func() {
		err := Disconnect()
		if err != nil {
			t.Fatalf("error disconnecting from db: %v", err)
		}
	}()
	result := db.Exec("DROP TABLE cards, users, decks;")
	if result.Error != nil {
		t.Fatalf("error deleting tables: %v", result.Error)
	}
}
