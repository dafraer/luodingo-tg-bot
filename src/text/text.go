package text

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Messages struct {
	Start             map[string]string `json:"start"`
	Help              map[string]string `json:"help"`
	ChooseDeck        map[string]string `json:"choose_deck"`
	ChooseCard        map[string]string `json:"choose_card"`
	ListDecks         map[string]string `json:"list_decks"`
	ListCards         map[string]string `json:"list_cards"`
	DeckDeleted       map[string]string `json:"deck_deleted"`
	ErrorCreatingDeck map[string]string `json:"error_creating_deck"`
	ErrorDeletingDeck map[string]string `json:"error_deleting_deck"`
	CardDeleted       map[string]string `json:"card_deleted"`
	DeckCreated       map[string]string `json:"deck_created"`
	CardCreated       map[string]string `json:"card_created"`
	ChooseDeckName    map[string]string `json:"choose_deck_name"`
	ChooseCardFront   map[string]string `json:"choose_card_front"`
	ChooseCardBack    map[string]string `json:"choose_card_back"`
	UnknownMessage    map[string]string `json:"unknown_message"`
	UnknownCommand    map[string]string `json:"unknown_command"`
	Stats             map[string]string `json:"stats"`
	NoDecks           map[string]string `json:"no_decks"`
	NoCards           map[string]string `json:"no_cards"`
	ErrorCreatingCard map[string]string `json:"error_creating_card"`
	CreateDeckFirst   map[string]string `json:"create_deck_first"`
	ShowAnswer        map[string]string `json:"show_answer"`
	StopStudy         map[string]string `json:"stop_study"`
	FinishedStudy     map[string]string `json:"finished_study"`
	DeckExists        map[string]string `json:"deck_exists"`
	TooLong           map[string]string `json:"too_long"`
	Done              map[string]string `json:"done"`
	Cancel            map[string]string `json:"cancel"`
	AddReverse        map[string]string `json:"add_reverse"`
	ReverseAdded      map[string]string `json:"reverse_added"`
}

func Load() *Messages {
	var msgs Messages

	text, err := os.Open("./src/text/messages.json")
	if err != nil {
		panic(fmt.Errorf("error opening text file: %v", err))
	}
	defer func() {
		if err := text.Close(); err != nil {
			panic(fmt.Errorf("error closing text file: %v", err))
		}
	}()

	var p []byte
	p, err = io.ReadAll(text)
	if err != nil {
		panic(fmt.Errorf("error reading english text file: %v", err))
	}

	if err = json.Unmarshal(p, &msgs); err != nil {
		panic(fmt.Errorf("error when parsing english text file, %v", err))
	}
	return &msgs
}
