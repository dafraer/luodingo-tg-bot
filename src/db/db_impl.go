package db

func CreateDeck(deck *Deck) (err error) {
	result := db.Exec("INSERT INTO decks (tg_user_id, name, cards_amount) VALUES (?, ?, ?);", deck.TgUserId, deck.Name, deck.CardsAmount)
	return result.Error
}

// GetDecks gets decks by user id
func GetDecks(userId int64) (decks []Deck, err error) {
	result := db.Raw("SELECT * FROM decks WHERE tg_user_id = ?;", userId).Scan(&decks)
	return decks, result.Error
}

// UpdateDeck updates deck's name
func UpdateDeck(oldDeck, newDeck *Deck) (err error) {
	result := db.Exec("UPDATE decks SET name = ? WHERE tg_user_id = ? AND name = ?;", newDeck.Name, oldDeck.TgUserId, oldDeck.Name)
	return result.Error
}

// DeleteDeck deletes deck based on deck's name and telegram id of tha user
func DeleteDeck(deck *Deck) (err error) {
	//Delete all card related to the deck
	result := db.Exec(`DELETE FROM cards USING decks WHERE decks.id = cards.deck_id AND decks.tg_user_id = ? AND decks.name = ?;`, deck.TgUserId, deck.Name)
	if result.Error != nil {
		return result.Error
	}

	//Delete the deck
	db.Exec("DELETE FROM decks WHERE tg_user_id = ? AND name = ?;", deck.TgUserId, deck.Name)
	return result.Error
}

// CreateCard creates card using deck name, tg_user_id and card struct
func CreateCard(deckName string, userId int64, card *Card) (id int, err error) {
	//Add card
	result := db.Exec("INSERT INTO cards (deck_id, front, back, learned) VALUES ((SELECT id FROM decks WHERE name = ? AND tg_user_id = ?), ?, ?, ?);", deckName, userId, card.Front, card.Back, card.Learned)
	if result.Error != nil {
		return -1, result.Error
	}

	//Add reverse card
	/*result = db.Exec("INSERT INTO cards (deck_id, front, back, learned) VALUES ((SELECT id FROM decks WHERE name = ? AND tg_user_id = ?), ?, ?, ?);", deckName, userId, card.Back, card.Front, card.Learned)
	if result.Error != nil {
		return result.Error
	}
	*/

	//Increment amount of cards in the deck
	result = db.Exec("UPDATE decks SET cards_amount = cards_amount + 1 WHERE name = ? AND tg_user_id = ?;", deckName, userId)
	if result.Error != nil {
		return -1, result.Error
	}

	result = db.Raw("SELECT cards.id FROM cards JOIN decks ON cards.deck_id = decks.id WHERE decks.tg_user_id = ? AND front = ? AND name = ?", userId, card.Front, deckName).Scan(&id)
	return id, result.Error
}

// GetCards returns card from a single deck based on deck name and tg_user_id
func GetCards(deckName string, userId int64) (cards []Card, err error) {
	result := db.Raw("SELECT cards.id, cards.deck_id, front, back, learned FROM cards JOIN decks ON decks.id = cards.deck_id WHERE decks.name = ? AND decks.tg_user_id = ?;", deckName, userId).Scan(&cards)
	return cards, result.Error
}

// GetUnlearnedCards returns only unlearned cards from a single deck based on deck name and tg_user_id
func GetUnlearnedCards(deckName string, userId int64) (cards []Card, err error) {
	result := db.Raw("SELECT cards.id, deck_id, front, back, learned FROM cards JOIN decks ON decks.id = cards.deck_id WHERE decks.name = ? AND decks.tg_user_id = ? AND learned = false;", deckName, userId).Scan(&cards)
	return cards, result.Error
}

// UpdateCard updates non nil fields of card struct
func UpdateCard(card *Card) (err error) {
	result := db.Table("cards").Where("id = ?", card.Id).Updates(card)
	return result.Error
}

// UnlearnCards sets 'learned' field of all cards in a single to deck to false
func UnlearnCards(deckName string, userId int64) (err error) {
	result := db.Exec("UPDATE cards SET learned = false FROM decks WHERE decks.id = cards.deck_id AND decks.name = ? AND decks.tg_user_id = ?;", deckName, userId)
	return result.Error
}

// DeleteCard deletes card by id, cards amount ion decremented based on deck name and tg_user_id
func DeleteCard(cardId, deckName string, userId int64) (err error) {
	result := db.Exec("DELETE FROM cards WHERE id = ?", cardId)
	if result.Error != nil {
		return result.Error
	}
	//Decrement amount of cards in the deck
	result = db.Exec("UPDATE decks SET cards_amount = cards_amount - 1 WHERE name = ? AND tg_user_id = ?;", deckName, userId)
	return result.Error
}

// GetCard return a specific card by card id
func GetCard(cardId int) (card Card, err error) {
	result := db.Table("cards").Where("id = ?", cardId).Scan(&card)
	return card, result.Error
}

// GetUser returns user based on tg_user_id
func GetUser(userId int64) (user *User, err error) {
	result := db.Raw("SELECT * FROM users WHERE tg_user_id = ?;", userId).Scan(&user)
	if result.RowsAffected == 0 {
		if err := CreateUser(&User{TgUserId: userId, State: 1, PageSelected: 1}); err != nil {
			return &User{}, err
		}
		return &User{TgUserId: userId, State: 1}, result.Error
	}
	return user, result.Error
}

// UpdateUser updates user based on non nil fields of User struct
func UpdateUser(user *User) (err error) {
	result := db.Table("users").Where("tg_user_id = ?", user.TgUserId).Updates(user)
	return result.Error
}

// CreateUser creates user based on struct
func CreateUser(user *User) (err error) {
	result := db.Table("users").Create(&user)
	return result.Error
}

// DeckExists checks if deck with the same name already exists
func DeckExists(d *Deck) (exists bool, err error) {
	var decks []int
	result := db.Raw("SELECT id  FROM decks WHERE tg_user_id = ? AND name = ?;", d.TgUserId, d.Name).Scan(&decks)
	if len(decks) == 0 {
		return false, result.Error
	}

	return true, result.Error
}

// CardExists checks if card with the same name already exists in the same deck
func CardExists(c *Card, deckName string, userId int64) (exists bool, err error) {
	var cards []int
	result := db.Raw("SELECT cards.id  FROM cards JOIN decks ON decks.id = deck_id WHERE name = ? AND front = ? AND tg_user_id = ?;", deckName, c.Front, userId).Scan(&cards)
	if len(cards) == 0 {
		return false, result.Error
	}
	return true, result.Error
}
