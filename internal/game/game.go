package game

import (
	"math/rand"
	"time"
)

type Suit string
type Rank string

const (
	Spades   Suit = "spades"
	Hearts   Suit = "hearts"
	Diamonds Suit = "diamonds"
	Clubs    Suit = "clubs"
)

var Suits = []Suit{Spades, Hearts, Diamonds, Clubs}
var Ranks = []Rank{"2", "3", "4", "5", "6", "7", "8", "9", "10", "jack", "queen", "king", "ace"}

type Card struct {
	Suit Suit `json:"suit"`
	Rank Rank `json:"rank"`
}

type Deck []Card
type Hand []Card

type GameState struct {
	Deck           Deck   `json:"deck"`
	PlayerHand     Hand   `json:"player_hand"`
	OpponentHand   Hand   `json:"opponent_hand"`
	Discarded      bool   `json:"discarded"`
	Showdown       bool   `json:"showdown"`
	PlayerSanity   int    `json:"player_sanity"`
	OpponentSanity int    `json:"opponent_sanity"`
	Pot            int    `json:"pot"`
	Turn           string `json:"turn"`
	GamePhase      string `json:"game_phase"` // "deal", "bet", "discard", "showdown", "complete"
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
	hand := append(Hand(nil), (*deck)[:count]...)
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
		Deck:           deck,
		PlayerHand:     player,
		OpponentHand:   opponent,
		Discarded:      false,
		Showdown:       false,
		PlayerSanity:   100,   // Starting sanity
		OpponentSanity: 100,   // Starting sanity
		Pot:            0,
		Turn:           "player",
		GamePhase:      "deal",
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
	g.GamePhase = "bet"
	return true
}

func (g *GameState) OpponentRespond() string {
	// Basic AI response - opponent matches the bet
	bet := g.Pot // Match current pot
	if bet > g.OpponentSanity {
		bet = g.OpponentSanity // All in
	}
	
	if bet == 0 {
		bet = 10 // Minimum bet
	}
	
	if g.OpponentSanity < bet {
		g.GamePhase = "complete"
		return "Opponent folds due to insufficient sanity"
	}
	
	g.OpponentSanity -= bet
	g.Pot += bet
	g.Turn = "player"
	g.GamePhase = "discard"
	return "Opponent calls and raises"
}

func (g *GameState) CanDiscard() bool {
	return g.GamePhase == "discard" && !g.Discarded
}

func (g *GameState) CanShowdown() bool {
	return g.Discarded || g.GamePhase == "showdown"
}

func (g *GameState) CompleteShowdown() {
	g.Showdown = true
	g.GamePhase = "complete"
	
	// Award pot to winner
	result := CompareHandsString(g.PlayerHand, g.OpponentHand)
	if result == "player" {
		g.PlayerSanity += g.Pot
	} else if result == "opponent" {
		g.OpponentSanity += g.Pot
	} else {
		// Split pot on tie
		g.PlayerSanity += g.Pot / 2
		g.OpponentSanity += g.Pot / 2
	}
	g.Pot = 0
}
