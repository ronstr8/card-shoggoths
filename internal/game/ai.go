package game

import (
	"math/rand"
	"os"
	"strconv"
)

// AIConfig holds configuration for the AI
type AIConfig struct {
	DiscardSimulations int
	Courage            float64 // Multiplier for winProb (e.g., 1.0 = normal, 1.2 = brave, 0.8 = timid)
}

var DefaultAI = AIConfig{
	DiscardSimulations: 100, // Number of trials per permutation
	Courage:            1.2, // Default courage (Brave)
}

func init() {
	if s := os.Getenv("AI_COURAGE"); s != "" {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			DefaultAI.Courage = v
		}
	}
}

// ScoreHand converts a HandValue into a comparable floating point score.
// This allows for averaging results from simulations.
// The scale is arbitrary but hierarchical: Rank >> Primary >> Secondary >> Kickers.
func ScoreHand(hv HandValue) float64 {
	score := float64(hv.Rank) * 1000000.0
	score += float64(hv.Primary) * 10000.0
	score += float64(hv.Secondary) * 100.0

	// Add kickers with diminishing weight
	for i, k := range hv.Kickers {
		divisor := 1.0
		for j := 0; j <= i; j++ {
			divisor *= 15.0 // Rank values go up to 14, so 15 is a safe base
		}
		score += float64(k) / divisor
	}
	return score
}

// getUnknownCards returns a deck containing all cards NOT in the exclusions list
func getUnknownCards(exclusions []Card) Deck {
	fullDeck := NewDeck() // This is already shuffled by NewDeck, effectively
	// But NewDeck shuffles. We need to filter.

	// Map for O(1) lookup
	excluded := make(map[string]bool)
	for _, c := range exclusions {
		excluded[string(c.Suit)+string(c.Rank)] = true
	}

	var unknown Deck
	for _, c := range fullDeck {
		if !excluded[string(c.Suit)+string(c.Rank)] {
			unknown = append(unknown, c)
		}
	}
	// Reshuffle to be sure, though NewDeck is shuffled.
	// The subset extraction preserves NewDeck's randomness order, so it's fine.
	return unknown
}

// ChooseDiscard determines the best indices to discard from the hand.
// It iterates through all 32 combinations of keeping/discarding cards.
func (c AIConfig) ChooseDiscard(hand Hand) []int {
	n := len(hand)
	limit := 1 << n // 2^n combinations

	var bestDiscards []int
	var maxAvgScore float64 = -1.0

	// Pre-calculate unknown deck to avoid recreating it every sim?
	// Actually, we must shuffle it every sim or draw randomly.
	baseUnknown := getUnknownCards(hand)

	for mask := 0; mask < limit; mask++ {
		// Identify kept cards and discard indices for this mask
		var kept Hand
		var discardIndices []int

		for i := 0; i < n; i++ {
			if (mask>>i)&1 == 0 {
				// 0 bit means keep
				kept = append(kept, hand[i])
			} else {
				// 1 bit means discard
				discardIndices = append(discardIndices, i)
			}
		}

		cardsNeeded := 5 - len(kept)
		if cardsNeeded == 0 {
			// No cards discarded, evaluate current hand
			val := EvaluateHand(kept)
			score := ScoreHand(val)
			if score > maxAvgScore {
				maxAvgScore = score
				bestDiscards = discardIndices
			}
			continue
		}

		if cardsNeeded > len(baseUnknown) {
			// Should not happen in standard play
			continue
		}

		// Run Monte Carlo simulations
		totalScore := 0.0

		// To speed up, we can copy the slice.
		// Ideally we just shuffle the indices of baseUnknown or pick k random.

		for sim := 0; sim < c.DiscardSimulations; sim++ {
			// Shuffle the unknown cards
			// Make a copy to shuffle? Or just pick random indices?
			// Fisher-Yates on a copy is safest and easiest.
			simDeck := make(Deck, len(baseUnknown))
			copy(simDeck, baseUnknown)

			rand.Shuffle(len(simDeck), func(i, j int) {
				simDeck[i], simDeck[j] = simDeck[j], simDeck[i]
			})

			// Draw needed cards
			drawn := simDeck[:cardsNeeded]

			// Form final hand
			finalHand := make(Hand, len(kept))
			copy(finalHand, kept)
			finalHand = append(finalHand, drawn...)

			val := EvaluateHand(finalHand)
			totalScore += ScoreHand(val)
		}

		avgScore := totalScore / float64(c.DiscardSimulations)

		if avgScore > maxAvgScore {
			maxAvgScore = avgScore
			bestDiscards = discardIndices
		}
	}

	// Ensure we preserve the original order reference or just return indices?
	// The indices [0, 2, 4] means discard hand[0], hand[2], hand[4].
	// This matches PerformDiscard expectation.

	return bestDiscards
}

// DecideAction determines the betting action.
// Returns action string ("fold", "check", "call", "raise") and amount (for raise).
func (c AIConfig) DecideAction(hand Hand, gameState *GameState) (string, int) {
	// 1. Evaluate current hand strength (0.0 to 1.0 roughly comparable to win rate)
	// Monte Carlo again? Or just raw hand value?
	// For betting, we want "Win Probability".
	// ScoreHand gives an absolute score, but checks nothing against opponent range.
	// A simple heuristic for 5-card draw:
	// - High pair or better is decent.
	// - Two pair is good.
	// - Three of a kind+ is strong.

	val := EvaluateHand(hand)
	// score := ScoreHand(val) // Unused for now, relying on rank-based winProb

	// Normalize score broadly?
	// High Card (Top A): ~14,000,000
	// One Pair (2s): ~202,000,000 (Rank 1 * 1mil) -> Wait, Rank HighCard=0.
	// One Pair=1. So 1,000,000 is base.
	// Two Pair=2. 2,000,000.

	// Pot Odds = Cost to Call / (Pot + Cost to Call)
	callAmount := gameState.CurrentBet - gameState.OpponentBet
	pot := gameState.Pot

	// Win Prob estimated
	var winProb float64
	switch val.Rank {
	case HighCard:
		winProb = 0.1
		if val.Kickers[0] > 10 {
			winProb = 0.2
		} // Face card high
	case OnePair:
		winProb = 0.4
		if val.Primary > 10 {
			winProb = 0.55
		} // High pair
	case TwoPair:
		winProb = 0.7
	case ThreeOfAKind:
		winProb = 0.85
	case Straight, Flush, FullHouse, FourOfAKind, StraightFlush, RoyalFlush:
		winProb = 0.95
	}

	// Apply Courage modifier
	winProb *= c.Courage
	if winProb > 0.99 {
		winProb = 0.99
	} // Cap at 99%

	// Adjust based on game phase if needed (bluffing?)
	// Simple logic for now.

	if callAmount == 0 {
		// Can check or bet
		if winProb > 0.6 {
			// Bet for value
			return "bet", 20 // Standard size
		}
		// Bluff chance?
		if rand.Float64() < 0.1*c.Courage {
			return "bet", 10
		}
		return "check", 0
	}

	// Facing a bet
	potOdds := float64(callAmount) / float64(pot+callAmount)

	// If winProb > potOdds, Call.
	// If winProb much higher, Raise.

	if winProb > potOdds+0.1 { // Margin of safety
		if winProb > 0.8 && rand.Float64() < 0.7*c.Courage {
			return "raise", 20
		}
		return "call", 0
	}

	// Bluff call?
	if rand.Float64() < 0.05*c.Courage {
		return "call", 0
	}

	return "fold", 0
}
