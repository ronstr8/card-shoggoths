package game

import (
	"math/rand"
	"time"
)

type Suit string
type Rank string

const (
	Spades   Suit = "Spades"
	Hearts   Suit = "Hearts"
	Diamonds Suit = "Diamonds"
	Clubs    Suit = "Clubs"
)

var Suits = []Suit{Spades, Hearts, Diamonds, Clubs}
var Ranks = []Rank{"2", "3", "4", "5", "6", "7", "8", "9", "10", "Jack", "Queen", "King", "Ace"}

type Card struct {
	Suit Suit `json:"suit"`
	Rank Rank `json:"rank"`
}

type Deck []Card

type Hand []Card

type GameState struct {
    Deck          Deck
    PlayerHand    Hand
    OpponentHand  Hand
    Discarded     bool
    Showdown      bool

    PlayerSanity   int
    OpponentSanity int
    Pot            int
    Turn           string // or maybe an enum like "player" / "opponent"
}


func NewDeck() Deck {
	var d Deck
	for _, s := range Suits {
		for _, r := range Ranks {
			d = append(d, Card{Suit: s, Rank: r})
		}
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(d), func(i, j int) { d[i], d[j] = d[j], d[i] })
	return d
}

func DealHand(deck *Deck, count int) Hand {
	hand := append(Hand(nil), (*deck)[:count]...);
	*deck = (*deck)[count:]
	return hand
}

func ReplaceCards(deck *Deck, hand *Hand, indices []int) {
	for _, i := range indices {
		(*hand)[i] = (*deck)[0]
		*deck = (*deck)[1:]
	}
}

func NewGame() *GameState {
	deck := NewDeck()
	player := DealHand(&deck, 5)
	opponent := DealHand(&deck, 5)
	return &GameState{
		Deck:         deck,
		PlayerHand:   player,
		OpponentHand: opponent,
		Discarded:    false,
		Showdown:     false,
	}
}


type Player struct {
	Name   string
	Hand   Hand
	Sanity int
}

func (g *GameState) PlaceBet(amount int) bool {
	if g.Turn != "player" || g.PlayerSanity < amount {
		return false
	}
	g.PlayerSanity -= amount
	g.Pot += amount
	g.Turn = "opponent"
	return true
}

func (g *GameState) OpponentRespond() string {
	// Basic AI response
	bet := 10
	if g.OpponentSanity < bet {
		return "Opponent folds"
	}
	g.OpponentSanity -= bet
	g.Pot += bet
	g.Turn = "showdown"
	return "Opponent calls"
}

