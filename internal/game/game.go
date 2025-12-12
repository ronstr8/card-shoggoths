package game

import (
	"fmt"
	"math/rand"
	"os"
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

type GamePhase int

const (
	PhaseAnte GamePhase = iota
	PhasePreDrawBetting
	PhaseDiscard
	PhasePostDrawBetting
	PhaseShowdown
	PhaseComplete
	PhaseGameOver
	PhaseESP
)

func (p GamePhase) String() string {
	switch p {
	case PhaseAnte:
		return "ante"
	case PhasePreDrawBetting:
		return "bet_pre"
	case PhaseDiscard:
		return "discard"
	case PhasePostDrawBetting:
		return "bet_post"
	case PhaseShowdown:
		return "showdown"
	case PhaseComplete:
		return "complete"
	case PhaseGameOver:
		return "game_over"
	case PhaseESP:
		return "esp"
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
	case "game_over":
		*p = PhaseGameOver
	case "esp":
		*p = PhaseESP
	default:
		return fmt.Errorf("unknown game phase: %s", string(text))
	}
	return nil
}

type GameState struct {
	ID          string        `json:"id"` // For persistence
	Deck        Deck          `json:"deck"`
	Players     []*Player     `json:"players"`
	RoundStates []*RoundState `json:"round_states"` // Transient state per player (index-matched)
	Pot         int           `json:"pot"`
	TurnIndex   int           `json:"turn_index"` // Index of player whose turn it is
	GamePhase   GamePhase     `json:"game_phase"`

	// Betting state
	CurrentBet   int    `json:"current_bet"`   // Amount to call
	LastAction   string `json:"last_action"`   // For UI display
	ActivePlayer string `json:"active_player"` // Legacy string for UI ("player" or "opponent") - kept for compatibility/easier UI mapping for now
	Winner       string `json:"winner"`        // Name of winner
	RevealOnFold bool   `json:"reveal_on_fold"`

	// ESP Minigame state
	ESP *ESPState `json:"esp,omitempty"`
}

// ESPState holds the state for the ESP training minigame
type ESPState struct {
	Hand1       Hand   `json:"hand1"`        // Top row (opponent's side)
	Hand2       Hand   `json:"hand2"`        // Bottom row (player's side)
	MatchIndex1 int    `json:"match_index1"` // Index in Hand1 of the matching card
	MatchIndex2 int    `json:"match_index2"` // Index in Hand2 of the matching card
	Attempts    int    `json:"attempts"`     // Number of guesses made
	Theme       string `json:"theme"`        // "primes", "faces", "odds", "evens"
	StartTime   int64  `json:"start_time"`   // Unix timestamp when ESP started
}

type Player struct {
	Name   string `json:"name"`
	IsAI   bool   `json:"is_ai"`
	Sanity int    `json:"sanity"`
}

type RoundState struct {
	Hand      Hand `json:"hand"`
	Bet       int  `json:"bet"` // Amount put in this round
	Folded    bool `json:"folded"`
	Discarded bool `json:"discarded"` // Has performed discard
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

func ShuffleDeck(deck *Deck) {
	rand.Shuffle(len(*deck), func(i, j int) { (*deck)[i], (*deck)[j] = (*deck)[j], (*deck)[i] })
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

	// Initialize Players (Identity)
	human := &Player{
		Name:   "You",
		IsAI:   false,
		Sanity: 100,
	}
	ai := &Player{
		Name:   "The Ancient One",
		IsAI:   true,
		Sanity: 100,
	}

	// Initialize Round States (Transient)
	humanState := &RoundState{Hand: []Card{}}
	aiState := &RoundState{Hand: []Card{}}

	return &GameState{
		Deck:         deck,
		Players:      []*Player{human, ai},
		RoundStates:  []*RoundState{humanState, aiState},
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
func (g *GameState) CollectAnte(amount int) bool {
	if g.GamePhase != PhaseAnte {
		return false
	}
	// Check sanity
	if g.Players[0].Sanity < amount {
		g.GamePhase = PhaseGameOver
		g.LastAction = fmt.Sprintf("%s has insufficient sanity for ante. Game Over.", g.Players[0].Name)
		return false
	}
	// AI Regeneration if bankrupt
	if g.Players[1].Sanity < amount {
		g.Players[1].Sanity = 100
		g.LastAction = "The Ancient One regenerates its form!"
	}

	// Deduct ante
	for _, p := range g.Players {
		p.Sanity -= amount
		g.Pot += amount
	}
	g.LastAction = fmt.Sprintf("Ante paid: %d", amount)

	// Deal cards
	for i, rs := range g.RoundStates {
		rs.Hand = DealHand(&g.Deck, 5)
		rs.Bet = 0
		rs.Folded = false
		rs.Discarded = false
		// Important: Ensure RoundStates aligns with Players
		_ = i
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

	for _, rs := range g.RoundStates {
		rs.Hand = []Card{}
		rs.Bet = 0
		rs.Folded = false
		rs.Discarded = false
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
	playerState := g.RoundStates[0]
	opponent := g.Players[1]
	opponentState := g.RoundStates[1]

	// Allow folding at any time
	if action == "fold" && g.GamePhase != PhaseComplete {
		g.GamePhase = PhaseComplete
		g.Winner = opponent.Name
		opponent.Sanity += g.Pot
		g.Pot = 0
		g.LastAction = fmt.Sprintf("You folded. %s wins.", opponent.Name)
		playerState.Folded = true
		return true, ""
	}

	// Strictly enforce turn: Index 0
	if g.TurnIndex != 0 || g.GamePhase == PhaseComplete {
		return false, "It is not your turn."
	}

	switch action {
	case "check":
		if g.CurrentBet > playerState.Bet {
			return false, "Cannot check when there is a bet to call."
		}
		g.LastAction = "You checked."
		g.TurnIndex = 1
		g.ActivePlayer = "opponent"
		return true, ""

	case "call":
		toCall := g.CurrentBet - playerState.Bet
		if toCall <= 0 {
			return false, "Nothing to call, please Check."
		}
		if player.Sanity < toCall {
			return false, "Not enough sanity to call."
		}
		player.Sanity -= toCall
		g.Pot += toCall
		playerState.Bet += toCall
		g.LastAction = "You called."

		// Round ends if action closes betting
		if opponentState.Bet == playerState.Bet {
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
		totalCost := (g.CurrentBet - playerState.Bet) + amount
		if player.Sanity < totalCost {
			return false, fmt.Sprintf("Not enough sanity. You need %d but have %d.", totalCost, player.Sanity)
		}

		player.Sanity -= totalCost
		g.Pot += totalCost
		playerState.Bet += totalCost
		g.CurrentBet = playerState.Bet

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
	// playerState := g.RoundStates[0] // Unused?
	opponent := g.Players[1]
	opponentState := g.RoundStates[1]

	// Use AI to decide action
	ai := DefaultAI
	// We need to support DecideAction possibly needing more context?
	// For now pass Hand and GameState.
	// NOTE: ai.DecideAction likely accesses g.Players[1].Bet - need to check AI!!
	action, amount := ai.DecideAction(opponentState.Hand, g)

	switch action {
	case "check":
		// Verify check is legal (no bet to call)
		if g.CurrentBet > opponentState.Bet {
			// Forced to call if AI made a mistake
			toCall := g.CurrentBet - opponentState.Bet
			if toCall > 0 {
				// Fallback to call logic
				if opponent.Sanity >= toCall {
					opponent.Sanity -= toCall
					g.Pot += toCall
					opponentState.Bet += toCall
					g.LastAction = fmt.Sprintf("%s calls.", opponent.Name)
					g.NextPhase()
				} else {
					// Fold
					g.GamePhase = PhaseComplete
					g.Winner = player.Name
					player.Sanity += g.Pot
					g.Pot = 0
					g.LastAction = fmt.Sprintf("%s folds (insufficient sanity).", opponent.Name)
					opponentState.Folded = true
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
		toCall := g.CurrentBet - opponentState.Bet
		if toCall > opponent.Sanity {
			// Fold
			g.GamePhase = PhaseComplete
			g.Winner = player.Name
			player.Sanity += g.Pot
			g.Pot = 0
			opponentState.Folded = true
			g.LastAction = fmt.Sprintf("%s folds.", opponent.Name)
			return
		}
		opponent.Sanity -= toCall
		g.Pot += toCall
		opponentState.Bet += toCall
		g.LastAction = fmt.Sprintf("%s calls.", opponent.Name)
		g.NextPhase()

	case "bet", "raise":
		totalCost := (g.CurrentBet - opponentState.Bet) + amount

		if opponent.Sanity < totalCost {
			// Fallback: Call if possible
			toCall := g.CurrentBet - opponentState.Bet
			if toCall > 0 && opponent.Sanity >= toCall {
				opponent.Sanity -= toCall
				g.Pot += toCall
				opponentState.Bet += toCall
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
				opponentState.Folded = true
				g.LastAction = fmt.Sprintf("%s folds.", opponent.Name)
			}
			return
		}

		opponent.Sanity -= totalCost
		g.Pot += totalCost
		opponentState.Bet += totalCost
		g.CurrentBet = opponentState.Bet

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
		opponentState.Folded = true
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
		for _, rs := range g.RoundStates {
			rs.Bet = 0
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
	playerState := g.RoundStates[0]
	// opponent := g.Players[1] // Unused in this func
	opponentState := g.RoundStates[1]

	// Player discards
	ReplaceCards(&g.Deck, &playerState.Hand, indices)
	playerState.Discarded = true

	// Opponent discards using AI
	ai := DefaultAI
	oppIndices := ai.ChooseDiscard(opponentState.Hand)

	ReplaceCards(&g.Deck, &opponentState.Hand, oppIndices)
	opponentState.Discarded = true

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
	playerState := g.RoundStates[0]
	opponent := g.Players[1]
	opponentState := g.RoundStates[1]

	result := CompareHands(playerState.Hand, opponentState.Hand)

	handRes := CompareHandsForDisplay(playerState.Hand, opponentState.Hand)
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

	// Immediate game over if player is bankrupt
	if player.Sanity <= 0 {
		g.GamePhase = PhaseGameOver
		g.LastAction = fmt.Sprintf("%s You lost everything. Game Over.", handRes.Message)
	}
}

// CanStartESP checks if ESP training is allowed in current phase
func (g *GameState) CanStartESP() bool {
	return g.GamePhase == PhaseComplete || g.GamePhase == PhaseAnte || g.GamePhase == PhaseGameOver
}

// ESP mystic themes - each uses a subset of ranks
var espThemes = map[string][]Rank{
	"primes": {"2", "3", "5", "7"},                 // Prime numbers
	"faces":  {"jack", "queen", "king", "ace"},     // Face cards
	"odds":   {"3", "5", "7", "9", "jack", "king"}, // Odd values
	"evens":  {"2", "4", "6", "8", "10", "queen"},  // Even values
}

var espThemeNames = []string{"primes", "faces", "odds", "evens"}
var espThemeMessages = map[string]string{
	"primes": "The primes align... 2, 3, 5, 7...",
	"faces":  "Royal visions emerge...",
	"odds":   "Odd energies swirl...",
	"evens":  "Even patterns crystallize...",
}

// StartESP initializes the ESP minigame with themed cards
func (g *GameState) StartESP() (bool, string) {
	if !g.CanStartESP() {
		return false, "The spirits are occupied. Complete your current hand first."
	}

	// Pick a random mystic theme
	theme := espThemeNames[rand.Intn(len(espThemeNames))]
	allowedRanks := espThemes[theme]

	// Build a deck with only allowed ranks
	var themedDeck Deck
	for _, s := range Suits {
		for _, r := range allowedRanks {
			themedDeck = append(themedDeck, Card{Suit: s, Rank: r})
		}
	}
	ShuffleDeck(&themedDeck)

	// Deal two hands of 5 cards each from themed deck
	hand1 := DealHand(&themedDeck, 5)
	hand2 := DealHand(&themedDeck, 5)

	// Guarantee at least one match by forcing a rank match
	matchIdx1 := rand.Intn(5)
	matchIdx2 := rand.Intn(5)

	// Make hand2[matchIdx2] have the same rank as hand1[matchIdx1]
	targetRank := hand1[matchIdx1].Rank
	for i, card := range themedDeck {
		if card.Rank == targetRank {
			hand2[matchIdx2] = card
			themedDeck = append(themedDeck[:i], themedDeck[i+1:]...)
			break
		}
	}

	g.ESP = &ESPState{
		Hand1:       hand1,
		Hand2:       hand2,
		MatchIndex1: matchIdx1,
		MatchIndex2: matchIdx2,
		Attempts:    0,
		Theme:       theme,
		StartTime:   time.Now().Unix(),
	}
	g.GamePhase = PhaseESP
	g.LastAction = espThemeMessages[theme] + " Find the matching cards!"
	return true, g.LastAction
}

// GuessESP checks if the player's guess is correct
// Returns (correct, message)
func (g *GameState) GuessESP(idx1, idx2 int) (bool, string) {
	if g.ESP == nil || g.GamePhase != PhaseESP {
		return false, "Not in ESP mode"
	}

	if idx1 < 0 || idx1 >= 5 || idx2 < 0 || idx2 >= 5 {
		return false, "Invalid card selection"
	}

	g.ESP.Attempts++

	// Check if the ranks match
	if g.ESP.Hand1[idx1].Rank == g.ESP.Hand2[idx2].Rank {
		// Correct!
		reward := 15
		g.Players[0].Sanity += reward
		g.LastAction = fmt.Sprintf("Your mind pierces the veil! +%d Sanity", reward)
		g.GamePhase = PhaseComplete
		g.ESP = nil
		return true, g.LastAction
	}

	// Wrong guess
	penalty := 5
	g.Players[0].Sanity -= penalty
	g.ESP.Attempts++

	if g.Players[0].Sanity <= 0 {
		g.GamePhase = PhaseGameOver
		g.LastAction = "The visions consumed you. Game Over."
		return false, g.LastAction
	}

	g.LastAction = fmt.Sprintf("The cards blur... -%d Sanity. Try again.", penalty)
	return false, g.LastAction
}

// ExitESP allows the player to leave ESP training early
func (g *GameState) ExitESP() {
	if g.GamePhase == PhaseESP {
		g.GamePhase = PhaseComplete
		g.ESP = nil
		g.LastAction = "You close your third eye."
	}
}
