package game

import (
	"fmt"
	"math/rand"
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

type GamePhase int

const (
	PhaseAnte GamePhase = iota
	PhasePreDrawBetting
	PhaseDiscard
	PhasePostDrawBetting
	PhaseShowdown
	PhaseComplete
)

func (p GamePhase) String() string {
	switch p {
	case PhaseAnte:
		return "ante"
	case PhasePreDrawBetting:
		return "pre_draw_betting"
	case PhaseDiscard:
		return "discard"
	case PhasePostDrawBetting:
		return "post_draw_betting"
	case PhaseShowdown:
		return "showdown"
	case PhaseComplete:
		return "complete"
	default:
		return "unknown"
	}
}

type GameState struct {
	Deck           Deck      `json:"deck"`
	PlayerHand     Hand      `json:"player_hand"`
	OpponentHand   Hand      `json:"opponent_hand"`
	Discarded      bool      `json:"discarded"`
	Showdown       bool      `json:"showdown"`
	PlayerSanity   int       `json:"player_sanity"`
	OpponentSanity int       `json:"opponent_sanity"`
	Pot            int       `json:"pot"`
	Turn           string    `json:"turn"`
	GamePhase      GamePhase `json:"game_phase"`

	// Betting state
	CurrentBet   int    `json:"current_bet"`   // Amount to call
	PlayerBet    int    `json:"player_bet"`    // Amount player has put effectively in this round
	OpponentBet  int    `json:"opponent_bet"`  // Amount opponent has put effectively in this round
	LastAction   string `json:"last_action"`   // For UI display
	ActivePlayer string `json:"active_player"` // "player" or "opponent"
	Winner       string `json:"winner"`        // Winner name if game over
}

func NewDeck() Deck {
	var d Deck
	for _, s := range Suits {
		for _, r := range Ranks {
			d = append(d, Card{Suit: s, Rank: r})
		}
	}
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
	// Hands are dealt after Ante

	return &GameState{
		Deck:           deck,
		PlayerHand:     []Card{},
		OpponentHand:   []Card{},
		Discarded:      false,
		Showdown:       false,
		PlayerSanity:   100, // Starting sanity
		OpponentSanity: 100, // Starting sanity
		Pot:            0,
		Turn:           "player",
		GamePhase:      PhaseAnte,
		CurrentBet:     0,
		PlayerBet:      0,
		OpponentBet:    0,
		LastAction:     "Game started. Ante up!",
		ActivePlayer:   "player",
	}
}

type Player struct {
	Name   string
	Hand   Hand
	Sanity int
}

// CollectAnte deducts ante and deals cards
func (g *GameState) CollectAnte(amount int) bool {
	if g.GamePhase != PhaseAnte {
		return false
	}
	if g.PlayerSanity < amount || g.OpponentSanity < amount {
		// Not enough sanity to play
		g.GamePhase = PhaseComplete
		g.LastAction = "Not enough sanity for ante."
		return false
	}

	g.PlayerSanity -= amount
	g.OpponentSanity -= amount
	g.Pot += amount * 2
	g.LastAction = fmt.Sprintf("Ante paid: %d", amount)

	// Deal cards
	g.PlayerHand = DealHand(&g.Deck, 5)
	g.OpponentHand = DealHand(&g.Deck, 5)

	// Transition to betting
	g.GamePhase = PhasePreDrawBetting
	g.CurrentBet = 0
	g.PlayerBet = 0
	g.OpponentBet = 0
	g.Turn = "player" // Player acts first
	g.ActivePlayer = "player"
	return true
}

func (g *GameState) PlayerAction(action string, amount int) bool {
	if g.Turn != "player" || g.GamePhase == PhaseComplete {
		return false
	}

	switch action {
	case "fold":
		g.GamePhase = PhaseComplete
		g.Winner = "opponent"
		g.OpponentSanity += g.Pot
		g.Pot = 0
		g.LastAction = "You folded. Opponent wins."
		return true

	case "check":
		if g.CurrentBet > g.PlayerBet {
			return false // Cannot check if there is a bet
		}
		g.LastAction = "You checked."
		g.Turn = "opponent"
		g.ActivePlayer = "opponent"
		// If both checked or logic needs check
		if g.ActivePlayer == "player" { // logic error check? no, wait for opponent
			// Opponent needs to act
		}
		return true

	case "call":
		toCall := g.CurrentBet - g.PlayerBet
		if toCall <= 0 {
			return false // Should check instead
		}
		if g.PlayerSanity < toCall {
			return false // Cannot afford
		}
		g.PlayerSanity -= toCall
		g.Pot += toCall
		g.PlayerBet += toCall
		g.LastAction = "You called."

		// Round ends if opponent was the aggressor
		if g.OpponentBet == g.PlayerBet {
			g.NextPhase()
		} else {
			g.Turn = "opponent"
			g.ActivePlayer = "opponent"
		}
		return true

	case "bet", "raise":
		// 'amount' is the TOTAL bet they want to be at, or the ADDITION?
		// Let's say amount is the RAISE amount ON TOP of current bet.
		// Or maybe easiest: amount is total contribution for this round?
		// Standard poker: Raise TO X.

		// Let's implement: Raise BY amount.
		// Total needed = (CurrentBet - PlayerBet) + Amount

		totalCost := (g.CurrentBet - g.PlayerBet) + amount
		if g.PlayerSanity < totalCost {
			return false
		}

		g.PlayerSanity -= totalCost
		g.Pot += totalCost
		g.PlayerBet += totalCost
		g.CurrentBet = g.PlayerBet

		g.LastAction = fmt.Sprintf("You raised by %d.", amount)
		g.Turn = "opponent"
		g.ActivePlayer = "opponent"
		return true
	}

	return false
}

func (g *GameState) OpponentTurn() {
	if g.Turn != "opponent" || g.GamePhase == PhaseComplete {
		return
	}

	// Simple AI
	// If check allowed, check.
	// If bet to call, call if decent hand or bluff.
	// Randomly raise.

	callAmount := g.CurrentBet - g.OpponentBet

	action := ""

	if callAmount == 0 {
		// Can check or bet
		roll := rand.Intn(100)
		if roll < 20 && g.OpponentSanity > 10 {
			// Bet/Raise
			raiseAmt := 10
			g.OpponentSanity -= raiseAmt
			g.Pot += raiseAmt
			g.OpponentBet += raiseAmt
			g.CurrentBet = g.OpponentBet
			action = fmt.Sprintf("Opponent bets %d.", raiseAmt)
			g.Turn = "player"
			g.ActivePlayer = "player"
		} else {
			// Check
			action = "Opponent checks."
			// Round complete?
			if g.ActivePlayer == "opponent" { // it was opp turn
				// If player checked first, then round over.
				// We need tracking of who started/acted.
				// Simplify: If opponent checks and he acted second (or first in generic flow), pass back or end.
				// In Heads Up: Dealer (Player) acts first Pre-Flop? Actually in Draw:
				// Pre-draw: Player 1st.
				// If Player Checks, Opp Checks -> Next Phase.
				// If Player Bets, Opp Calls -> Next Phase.
				g.NextPhase()
			}
		}
	} else {
		// Must call or fold (or raise)
		if callAmount > g.OpponentSanity {
			// Fold
			g.GamePhase = PhaseComplete
			g.Winner = "player"
			g.PlayerSanity += g.Pot
			g.Pot = 0
			g.LastAction = "Opponent folds. You win!"
			return
		}

		// Call
		g.OpponentSanity -= callAmount
		g.Pot += callAmount
		g.OpponentBet += callAmount
		action = "Opponent calls."
		g.NextPhase()
	}

	if g.GamePhase != PhaseComplete {
		g.LastAction = action
	}
}

func (g *GameState) NextPhase() {
	// Transition logic
	switch g.GamePhase {
	case PhasePreDrawBetting:
		g.GamePhase = PhaseDiscard
		g.Turn = "player"
		g.ActivePlayer = "player"
		g.LastAction = "Betting complete. Choose cards to discard."
		// Reset bets for next round
		g.CurrentBet = 0
		g.PlayerBet = 0
		g.OpponentBet = 0

	case PhaseDiscard:
		// After discard (assuming wait for both, but simplified: Player discards in UI, then immediately Opponent discards and we toggle phase)
		// Wait, we need a distinct transitions.
		// Let's assume this method is called after both have discarded.
		g.GamePhase = PhasePostDrawBetting
		g.Turn = "player"
		g.ActivePlayer = "player"
		g.LastAction = "Cards exchanged. Final betting round."

	case PhasePostDrawBetting:
		g.GamePhase = PhaseShowdown
		g.CompleteShowdown()
	}
}

func (g *GameState) PerformDiscard(indices []int) {
	if g.GamePhase != PhaseDiscard {
		return
	}
	// Player discards
	ReplaceCards(&g.Deck, &g.PlayerHand, indices)

	// Opponent discards (simple AI: discard non-paired low cards)
	// TODO: Better AI for discard
	// For now, discard random 0-2 cards
	discardCount := rand.Intn(3)
	var oppIndices []int
	for i := 0; i < discardCount; i++ {
		oppIndices = append(oppIndices, i) // simple discard first few
	}
	ReplaceCards(&g.Deck, &g.OpponentHand, oppIndices)

	g.Discarded = true // Mark as done
	g.NextPhase()
}

func (g *GameState) CompleteShowdown() {
	g.Showdown = true
	g.GamePhase = PhaseComplete

	result := CompareHands(g.PlayerHand, g.OpponentHand)

	handRes := CompareHandsForDisplay(g.PlayerHand, g.OpponentHand)
	g.LastAction = handRes.Message

	switch result {
	case ResultHand1Wins:
		g.PlayerSanity += g.Pot
		g.Winner = "player"
	case ResultHand2Wins:
		g.OpponentSanity += g.Pot
		g.Winner = "opponent"
	default:
		g.PlayerSanity += g.Pot / 2
		g.OpponentSanity += g.Pot / 2
		g.Winner = "tie"
	}
	g.Pot = 0
}
