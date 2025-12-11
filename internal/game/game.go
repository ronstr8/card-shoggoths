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
		return "ante" // Was "deal", changing to "ante" for clarity? Or keep "deal"? "ante" matches PhaseAnte.
	case PhasePreDrawBetting:
		return "bet_pre"
	case PhaseDiscard:
		return "discard"
	case PhasePostDrawBetting:
		return "bet_post"
	case PhaseShowdown:
		return "showdown"
	case PhaseComplete:
		return "complete" // Was "end"
	default:
		return "unknown"
	}
}

func (p GamePhase) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *GamePhase) UnmarshalText(text []byte) error {
	switch string(text) {
	case "ante", "deal": // Support legacy "deal"
		*p = PhaseAnte
	case "bet_pre":
		*p = PhasePreDrawBetting
	case "bet": // Unambiguous legacy mapping -> PreDraw
		*p = PhasePreDrawBetting
	case "discard":
		*p = PhaseDiscard
	case "bet_post":
		*p = PhasePostDrawBetting
	case "showdown":
		*p = PhaseShowdown
	case "complete", "end": // Support legacy "end"
		*p = PhaseComplete
	default:
		return fmt.Errorf("unknown game phase: %s", string(text))
	}
	return nil
}

type GameState struct {
	ID        string    `json:"id"` // For persistence
	Deck      Deck      `json:"deck"`
	Players   []*Player `json:"players"`
	Pot       int       `json:"pot"`
	TurnIndex int       `json:"turn_index"` // Index of player whose turn it is
	GamePhase GamePhase `json:"game_phase"`

	// Betting state
	CurrentBet   int    `json:"current_bet"`   // Amount to call
	LastAction   string `json:"last_action"`   // For UI display
	ActivePlayer string `json:"active_player"` // Legacy string for UI ("player" or "opponent") - kept for compatibility/easier UI mapping for now
	Winner       string `json:"winner"`        // Name of winner
	RevealOnFold bool   `json:"reveal_on_fold"`
}

type Player struct {
	Name      string `json:"name"`
	IsAI      bool   `json:"is_ai"`
	Hand      Hand   `json:"hand"`
	Sanity    int    `json:"sanity"`
	Bet       int    `json:"bet"` // Amount put in this round
	Folded    bool   `json:"folded"`
	Discarded bool   `json:"discarded"` // Has performed discard
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
	if len(*deck) < count {
		return Hand{} // Should handle error
	}
	hand := append(Hand(nil), (*deck)[:count]...)
	*deck = (*deck)[count:]
	return hand
}

func ReplaceCards(deck *Deck, hand *Hand, indices []int) {
	for _, i := range indices {
		if i >= 0 && i < len(*hand) && len(*deck) > 0 {
			(*hand)[i] = (*deck)[0]
			*deck = (*deck)[1:]
		}
	}
}

func NewGame() *GameState {
	deck := NewDeck()

	// Initialize Players
	human := &Player{
		Name:   "You",
		IsAI:   false,
		Sanity: 100,
		Hand:   []Card{},
	}
	ai := &Player{
		Name:   "The Ancient One",
		IsAI:   true,
		Sanity: 100,
		Hand:   []Card{},
	}

	return &GameState{
		Deck:         deck,
		Players:      []*Player{human, ai},
		TurnIndex:    0, // Human starts
		GamePhase:    PhaseAnte,
		Pot:          0,
		CurrentBet:   0,
		LastAction:   "Game started. Ante up!",
		ActivePlayer: "player",
		RevealOnFold: GlobalRevealOnFold,
	}
}

// CollectAnte deducts ante and deals cards
// CollectAnte deducts ante and deals cards
func (g *GameState) CollectAnte(amount int) bool {
	if g.GamePhase != PhaseAnte {
		return false
	}
	// Check sanity
	for _, p := range g.Players {
		if p.Sanity < amount {
			g.GamePhase = PhaseComplete
			g.LastAction = fmt.Sprintf("%s has not enough sanity for ante.", p.Name)
			return false
		}
	}

	// Deduct ante
	for _, p := range g.Players {
		p.Sanity -= amount
		g.Pot += amount
	}
	g.LastAction = fmt.Sprintf("Ante paid: %d", amount)

	// Deal cards
	for _, p := range g.Players {
		p.Hand = DealHand(&g.Deck, 5)
		p.Bet = 0
		p.Folded = false
		p.Discarded = false
	}

	// Transition to betting
	g.GamePhase = PhasePreDrawBetting
	g.CurrentBet = 0

	// Reset turn to 0 (Human)
	g.TurnIndex = 0
	g.ActivePlayer = "player" // Keep legacy sync
	return true
}

func (g *GameState) NewRound() {
	deck := NewDeck()
	g.Deck = deck

	for _, p := range g.Players {
		p.Hand = []Card{}
		p.Bet = 0
		p.Folded = false
		p.Discarded = false
	}

	g.Pot = 0
	g.TurnIndex = 0
	g.GamePhase = PhaseAnte
	g.CurrentBet = 0
	g.LastAction = "New round started. Ante up!"
	g.ActivePlayer = "player"
	g.Winner = ""
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
	// Security check: only allow actions for player 0 (human) via this method?
	// The handler calls this. Let's assume this method is specifically for the Human Player (index 0).
	player := g.Players[0]
	opponent := g.Players[1]

	// Allow folding at any time
	if action == "fold" && g.GamePhase != PhaseComplete {
		g.GamePhase = PhaseComplete
		g.Winner = opponent.Name
		opponent.Sanity += g.Pot
		g.Pot = 0
		g.LastAction = fmt.Sprintf("You folded. %s wins.", opponent.Name)
		player.Folded = true
		return true, ""
	}

	// Strictly enforce turn: Index 0
	if g.TurnIndex != 0 || g.GamePhase == PhaseComplete {
		return false, "It is not your turn."
	}

	switch action {
	case "check":
		if g.CurrentBet > player.Bet {
			return false, "Cannot check when there is a bet to call."
		}
		g.LastAction = "You checked."
		g.TurnIndex = 1
		g.ActivePlayer = "opponent"
		return true, ""

	case "call":
		toCall := g.CurrentBet - player.Bet
		if toCall <= 0 {
			return false, "Nothing to call, please Check."
		}
		if player.Sanity < toCall {
			return false, "Not enough sanity to call."
		}
		player.Sanity -= toCall
		g.Pot += toCall
		player.Bet += toCall
		g.LastAction = "You called."

		// Round ends if action closes betting
		if opponent.Bet == player.Bet {
			g.NextPhase()
		} else {
			g.TurnIndex = 1
			g.ActivePlayer = "opponent"
		}
		return true, ""

	case "bet", "raise":
		if amount <= 0 {
			return false, "Bet amount must be positive."
		}
		totalCost := (g.CurrentBet - player.Bet) + amount
		if player.Sanity < totalCost {
			return false, fmt.Sprintf("Not enough sanity. You need %d but have %d.", totalCost, player.Sanity)
		}

		player.Sanity -= totalCost
		g.Pot += totalCost
		player.Bet += totalCost
		g.CurrentBet = player.Bet

		if action == "bet" {
			g.LastAction = fmt.Sprintf("You bet %d.", amount)
		} else {
			g.LastAction = fmt.Sprintf("You raised by %d.", amount)
		}
		g.TurnIndex = 1
		g.ActivePlayer = "opponent"
		return true, ""
	}

	return false, "Invalid action."
}

func (g *GameState) OpponentTurn() {
	if g.TurnIndex != 1 || g.GamePhase == PhaseComplete {
		return
	}

	player := g.Players[0]
	opponent := g.Players[1]

	// Use AI to decide action
	ai := DefaultAI
	// We need to support DecideAction possibly needing more context?
	// For now pass Hand and GameState.
	action, amount := ai.DecideAction(opponent.Hand, g)

	switch action {
	case "check":
		// Verify check is legal (no bet to call)
		if g.CurrentBet > opponent.Bet {
			// Forced to call if AI made a mistake
			toCall := g.CurrentBet - opponent.Bet
			if toCall > 0 {
				// Fallback to call logic
				if opponent.Sanity >= toCall {
					opponent.Sanity -= toCall
					g.Pot += toCall
					opponent.Bet += toCall
					g.LastAction = fmt.Sprintf("%s calls.", opponent.Name)
					g.NextPhase()
				} else {
					// Fold
					g.GamePhase = PhaseComplete
					g.Winner = player.Name
					player.Sanity += g.Pot
					g.Pot = 0
					g.LastAction = fmt.Sprintf("%s folds (insufficient sanity).", opponent.Name)
					opponent.Folded = true
				}
				return
			}
		}

		g.LastAction = fmt.Sprintf("%s checks.", opponent.Name)
		if g.ActivePlayer == "opponent" {
			// If opponent acted second/last in sequence, usually NextPhase?
			// Simplistic turn handling: If P0 Checked, P1 Checks -> Next
			g.NextPhase()
		} else {
			// If P1 (Opponent) is acting, give turn back to P0
			g.TurnIndex = 0
			g.ActivePlayer = "player"
		}

	case "call":
		toCall := g.CurrentBet - opponent.Bet
		if toCall > opponent.Sanity {
			// Fold
			g.GamePhase = PhaseComplete
			g.Winner = player.Name
			player.Sanity += g.Pot
			g.Pot = 0
			opponent.Folded = true
			g.LastAction = fmt.Sprintf("%s folds.", opponent.Name)
			return
		}
		opponent.Sanity -= toCall
		g.Pot += toCall
		opponent.Bet += toCall
		g.LastAction = fmt.Sprintf("%s calls.", opponent.Name)
		g.NextPhase()

	case "bet", "raise":
		totalCost := (g.CurrentBet - opponent.Bet) + amount

		if opponent.Sanity < totalCost {
			// Fallback: Call if possible
			toCall := g.CurrentBet - opponent.Bet
			if toCall > 0 && opponent.Sanity >= toCall {
				opponent.Sanity -= toCall
				g.Pot += toCall
				opponent.Bet += toCall
				g.LastAction = fmt.Sprintf("%s calls.", opponent.Name)
				g.NextPhase()
			} else if toCall == 0 {
				g.LastAction = fmt.Sprintf("%s checks.", opponent.Name)
				g.TurnIndex = 0
				g.ActivePlayer = "player"
			} else {
				// Fold
				g.GamePhase = PhaseComplete
				g.Winner = player.Name
				player.Sanity += g.Pot
				g.Pot = 0
				opponent.Folded = true
				g.LastAction = fmt.Sprintf("%s folds.", opponent.Name)
			}
			return
		}

		opponent.Sanity -= totalCost
		g.Pot += totalCost
		opponent.Bet += totalCost
		g.CurrentBet = opponent.Bet

		if action == "bet" {
			g.LastAction = fmt.Sprintf("%s bets %d.", opponent.Name, amount)
		} else {
			g.LastAction = fmt.Sprintf("%s raises by %d.", opponent.Name, amount)
		}
		g.TurnIndex = 0
		g.ActivePlayer = "player"

	case "fold":
		g.GamePhase = PhaseComplete
		g.Winner = player.Name
		player.Sanity += g.Pot
		g.Pot = 0
		opponent.Folded = true
		g.LastAction = fmt.Sprintf("%s folds. You win!", opponent.Name)
	}
}

func (g *GameState) NextPhase() {
	// Transition logic
	switch g.GamePhase {
	case PhasePreDrawBetting:
		g.GamePhase = PhaseDiscard
		g.TurnIndex = 0
		g.ActivePlayer = "player"
		g.LastAction = "Betting complete. Choose cards to discard."
		// Reset bets for next round
		g.CurrentBet = 0
		for _, p := range g.Players {
			p.Bet = 0
		}

	case PhaseDiscard:
		g.GamePhase = PhasePostDrawBetting
		g.TurnIndex = 0
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
	player := g.Players[0]
	opponent := g.Players[1]

	// Player discards
	ReplaceCards(&g.Deck, &player.Hand, indices)
	player.Discarded = true

	// Opponent discards using AI
	ai := DefaultAI
	oppIndices := ai.ChooseDiscard(opponent.Hand)

	ReplaceCards(&g.Deck, &opponent.Hand, oppIndices)
	opponent.Discarded = true

	// If both done (which they are), next phase
	g.NextPhase()
}

func (g *GameState) CanDiscard() bool {
	return g.GamePhase == PhaseDiscard
}

func (g *GameState) CanShowdown() bool {
	return g.GamePhase == PhaseShowdown || g.GamePhase == PhaseComplete
}

func (g *GameState) CompleteShowdown() {
	// g.Showdown field meant "is showdown happening/visible"?
	// We'll leave it for UI compatibility, but phase is Complete
	g.GamePhase = PhaseComplete

	player := g.Players[0]
	opponent := g.Players[1]

	result := CompareHands(player.Hand, opponent.Hand)

	handRes := CompareHandsForDisplay(player.Hand, opponent.Hand)
	g.LastAction = handRes.Message

	switch result {
	case ResultHand1Wins:
		player.Sanity += g.Pot
		g.Winner = player.Name
	case ResultHand2Wins:
		opponent.Sanity += g.Pot
		g.Winner = opponent.Name
	default:
		player.Sanity += g.Pot / 2
		opponent.Sanity += g.Pot / 2
		g.Winner = "tie"
	}
	g.Pot = 0
}
