package text

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Messages struct {
	Start             string
	Help              string
	ChooseDeck        string
	ChooseCard        string
	ListDecks         string
	ListCards         string
	DeckDeleted       string
	ErrorCreatingDeck string
	ErrorDeletingDeck string
	CardDeleted       string
	DeckCreated       string
	CardCreated       string
	ChooseDeckName    string
	ChooseCardFront   string
	ChooseCardBack    string
	UnknownMessage    string
	UnknownCommand    string
	Stats             string
	NoDecks           string
	ErrorCreatingCard string
	CreateDeckFirst   string
}

func LoadEnMessages() Messages {
	var en Messages
	enText, err := os.Open("./src/text/en.json")
	if err != nil {
		panic(fmt.Errorf("error opening english text file: %v", err))
	}
	defer func() {
		if err := enText.Close(); err != nil {
			panic(fmt.Errorf("error closing english text file: %v", err))
		}
	}()

	var p []byte
	p, err = io.ReadAll(enText)
	if err != nil {
		panic(fmt.Errorf("error reading english text file: %v", err))
	}
	if err = json.Unmarshal(p, &en); err != nil {
		panic(fmt.Errorf("error when parsing english text file, %v", err))
	}
	return en
}

func LoadRuMessages() Messages {
	var ru Messages
	ruText, err := os.Open("./src/text/ru.json")
	if err != nil {
		panic(fmt.Errorf("error opening russian text file, %v", err))
	}
	defer func() {
		if err := ruText.Close(); err != nil {
			panic(fmt.Errorf("error closing russian text file: %v", err))
		}
	}()

	p, err := io.ReadAll(ruText)
	if err != nil {
		panic(fmt.Errorf("error opening russian text file, %v", err))
	}
	if err = json.Unmarshal(p, &ru); err != nil {
		panic(fmt.Errorf("error parsing russian text file, %v", err))
	}
	return ru
}
