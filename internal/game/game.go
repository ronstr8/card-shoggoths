package game

import (
	"fmt"
	"math/rand"
	"os"
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
		return "deal"
	case PhasePreDrawBetting:
		return "bet" // Mapped to 'bet' for frontend
	case PhaseDiscard:
		return "discard"
	case PhasePostDrawBetting:
		return "bet" // Reuse 'bet' for frontend
	case PhaseShowdown:
		return "showdown"
	case PhaseComplete:
		return "end"
	default:
		return "unknown"
	}
}

func (p GamePhase) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
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
	CurrentBet   int    `json:"current_bet"`    // Amount to call
	PlayerBet    int    `json:"player_bet"`     // Amount player has put effectively in this round
	OpponentBet  int    `json:"opponent_bet"`   // Amount opponent has put effectively in this round
	LastAction   string `json:"last_action"`    // For UI display
	ActivePlayer string `json:"active_player"`  // "player" or "opponent"
	Winner       string `json:"winner"`         // Winner name if game over
	RevealOnFold bool   `json:"reveal_on_fold"` // Config flag
	OpponentName string `json:"opponent_name"`  // Name of the opponent
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
		RevealOnFold:   GlobalRevealOnFold,
		OpponentName:   "The Ancient One",
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

func (g *GameState) NewRound() {
	deck := NewDeck()

	// Reset round-specific state
	g.Deck = deck
	g.PlayerHand = []Card{}
	g.OpponentHand = []Card{}
	g.Discarded = false
	g.Showdown = false
	g.Pot = 0
	g.Turn = "player"
	g.GamePhase = PhaseAnte
	g.CurrentBet = 0
	g.PlayerBet = 0
	g.OpponentBet = 0
	g.LastAction = "New round started. Ante up!"
	g.ActivePlayer = "player"
	g.RevealOnFold = GlobalRevealOnFold
}

var GlobalRevealOnFold = true

func init() {
	if val := os.Getenv("REVEAL_ON_FOLD"); val == "true" || val == "1" {
		GlobalRevealOnFold = true
	}
}

// ... (existing code)

func (g *GameState) PlayerAction(action string, amount int) (bool, string) {
	// Allow folding at any time as a "Force Quit" / Resign, provided game is active
	if action == "fold" && g.GamePhase != PhaseComplete {
		g.GamePhase = PhaseComplete
		g.Winner = "opponent"
		g.OpponentSanity += g.Pot
		g.Pot = 0
		g.LastAction = fmt.Sprintf("You folded. %s wins.", g.OpponentName)
		g.Showdown = true // Ensure reveal logic works if enabled
		return true, ""
	}

	if g.Turn != "player" || g.GamePhase == PhaseComplete {
		return false, "It is not your turn."
	}

	switch action {
	// "fold" handled above as special case

	case "check":
		if g.CurrentBet > g.PlayerBet {
			return false, "Cannot check when there is a bet to call."
		}
		g.LastAction = "You checked."
		g.Turn = "opponent"
		g.ActivePlayer = "opponent"
		return true, ""

	case "call":
		toCall := g.CurrentBet - g.PlayerBet
		if toCall <= 0 {
			// Actually, calling 0 IS a check. But let's separate them or allow it?
			// UI should handle "Check" button visibility.
			// But if user hits "Call" on 0 bet, treat as Check?
			// Logic below says false. Let's return error "Nothing to call, use Check".
			return false, "Nothing to call, please Check."
		}
		if g.PlayerSanity < toCall {
			return false, "Not enough sanity to call."
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
		return true, ""

	case "bet", "raise":
		// Raise BY amount logic
		if amount <= 0 {
			return false, "Bet amount must be positive."
		}

		totalCost := (g.CurrentBet - g.PlayerBet) + amount
		if g.PlayerSanity < totalCost {
			return false, fmt.Sprintf("Not enough sanity. You need %d but have %d.", totalCost, g.PlayerSanity)
		}

		g.PlayerSanity -= totalCost
		g.Pot += totalCost
		g.PlayerBet += totalCost
		g.CurrentBet = g.PlayerBet

		if action == "bet" {
			g.LastAction = fmt.Sprintf("You bet %d.", amount)
		} else {
			g.LastAction = fmt.Sprintf("You raised by %d.", amount)
		}
		g.Turn = "opponent"
		g.ActivePlayer = "opponent"
		return true, ""
	}

	return false, "Invalid action."
}

func (g *GameState) OpponentTurn() {
	if g.Turn != "opponent" || g.GamePhase == PhaseComplete {
		return
	}

	// Use AI to decide action
	ai := DefaultAI
	action, amount := ai.DecideAction(g.OpponentHand, g)

	switch action {
	case "check":
		// Verify check is legal (no bet to call)
		if g.CurrentBet > g.OpponentBet {
			// Forced to call (or fold, but AI said check which implies it thinks it can?
			// Actually DecideAction logic handles call vs check based on input.
			// But let's be safe. If we must call but tried to check, just call.
			toCall := g.CurrentBet - g.OpponentBet
			if toCall > 0 {
				// Fallback to call logic
				if g.OpponentSanity >= toCall {
					g.OpponentSanity -= toCall
					g.Pot += toCall
					g.OpponentBet += toCall
					g.LastAction = fmt.Sprintf("%s calls.", g.OpponentName)
					g.NextPhase()
				} else {
					// Fold
					g.GamePhase = PhaseComplete
					g.Winner = "player"
					g.PlayerSanity += g.Pot
					g.Pot = 0
					g.LastAction = fmt.Sprintf("%s folds (insufficient sanity).", g.OpponentName)
				}
				return
			}
		}

		g.LastAction = fmt.Sprintf("%s checks.", g.OpponentName)
		if g.ActivePlayer == "opponent" {
			// If opponent acted second/last, round over
			g.NextPhase()
		} else {
			// Player's turn next?
			// "check" usually passes action.
			// If Player Checked, Opponent Checks -> Round End.
			// If Player Bet, Opponent Check -> Illegal, handled above.
			// If new round (Opponent First?), Opp Check -> Player Turn.

			// In this engine, who goes first?
			// g.Turn = "player"?
			// Let's assume generic state flip.
			g.Turn = "player"
			g.ActivePlayer = "player"
		}

	case "call":
		toCall := g.CurrentBet - g.OpponentBet
		if toCall > g.OpponentSanity {
			// Fold
			g.GamePhase = PhaseComplete
			g.Winner = "player"
			g.PlayerSanity += g.Pot
			g.Pot = 0
			g.LastAction = "Opponent folds."
			return
		}
		g.OpponentSanity -= toCall
		g.Pot += toCall
		g.OpponentBet += toCall
		g.LastAction = fmt.Sprintf("%s calls.", g.OpponentName)
		g.NextPhase()

	case "bet", "raise":
		// Raise amount
		// amount returned by AI is 'raise by'
		totalCost := (g.CurrentBet - g.OpponentBet) + amount

		if g.OpponentSanity < totalCost {
			// Just call/check if can't afford raise
			// Or all-in? Let's just call/check for simplicity to avoid side-pot logic
			toCall := g.CurrentBet - g.OpponentBet
			if toCall > 0 && g.OpponentSanity >= toCall {
				g.OpponentSanity -= toCall
				g.Pot += toCall
				g.OpponentBet += toCall
				g.LastAction = fmt.Sprintf("%s calls.", g.OpponentName)
				g.NextPhase()
			} else if toCall == 0 {
				g.LastAction = fmt.Sprintf("%s checks.", g.OpponentName)
				g.Turn = "player"
				g.ActivePlayer = "player"
			} else {
				// Fold
				g.GamePhase = PhaseComplete
				g.Winner = "player"
				g.PlayerSanity += g.Pot
				g.Pot = 0
				g.LastAction = fmt.Sprintf("%s folds.", g.OpponentName)
			}
			return
		}

		g.OpponentSanity -= totalCost
		g.Pot += totalCost
		g.OpponentBet += totalCost
		g.CurrentBet = g.OpponentBet

		if action == "bet" {
			g.LastAction = fmt.Sprintf("%s bets %d.", g.OpponentName, amount)
		} else {
			g.LastAction = fmt.Sprintf("%s raises by %d.", g.OpponentName, amount)
		}
		g.Turn = "player"
		g.ActivePlayer = "player"

	case "fold":
		g.GamePhase = PhaseComplete
		g.Winner = "player"
		g.PlayerSanity += g.Pot
		g.Pot = 0
		g.LastAction = fmt.Sprintf("%s folds. You win!", g.OpponentName)
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

	// Opponent discards using AI
	ai := DefaultAI
	oppIndices := ai.ChooseDiscard(g.OpponentHand)

	ReplaceCards(&g.Deck, &g.OpponentHand, oppIndices)

	g.Discarded = true // Mark as done
	g.NextPhase()
}

func (g *GameState) CanDiscard() bool {
	return g.GamePhase == PhaseDiscard
}

func (g *GameState) CanShowdown() bool {
	return g.GamePhase == PhaseShowdown || g.GamePhase == PhaseComplete
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
