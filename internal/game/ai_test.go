package game

import (
	"fmt"
	"testing"
)

func TestScoreHand(t *testing.T) {
	// Sanity check that higher ranks get higher scores
	h1 := []Card{{Hearts, "ace"}, {Spades, "ace"}, {Clubs, "5"}, {Diamonds, "2"}, {Hearts, "3"}}   // Pair of Aces
	h2 := []Card{{Hearts, "king"}, {Spades, "king"}, {Clubs, "5"}, {Diamonds, "2"}, {Hearts, "3"}} // Pair of Kings
	h3 := []Card{{Hearts, "2"}, {Spades, "3"}, {Clubs, "4"}, {Diamonds, "5"}, {Hearts, "7"}}       // High Card

	s1 := ScoreHand(EvaluateHand(h1))
	s2 := ScoreHand(EvaluateHand(h2))
	s3 := ScoreHand(EvaluateHand(h3))

	if s1 <= s2 {
		t.Errorf("Pair of Aces should score higher than Pair of Kings: %f vs %f", s1, s2)
	}
	if s2 <= s3 {
		t.Errorf("Pair should score higher than High Card: %f vs %f", s2, s3)
	}
}

func TestChooseDiscard(t *testing.T) {
	ai := DefaultAI
	ai.DiscardSimulations = 200 // usage for test

	// Case 1: 4 to a Royal Flush. Should discard the junk card.
	// Hand: A, K, Q, J of Hearts, and a 2 of Clubs.
	hand1 := Hand{
		{Hearts, "ace"},
		{Hearts, "king"},
		{Hearts, "queen"},
		{Hearts, "jack"},
		{Clubs, "2"},
	}

	discards1 := ai.ChooseDiscard(hand1)
	if len(discards1) != 1 {
		t.Fatalf("Expected 1 discard, got %d", len(discards1))
	}
	// The 2 of clubs is at index 4
	if discards1[0] != 4 {
		t.Errorf("Expected to discard index 4 (2 of clubs), got %d", discards1[0])
	}
}

func TestChooseDiscard_KeepPatHand(t *testing.T) {
	ai := DefaultAI
	ai.DiscardSimulations = 50

	// Case 2: Full House, keep all.
	hand2 := Hand{
		{Hearts, "ace"}, {Spades, "ace"}, {Clubs, "ace"},
		{Hearts, "king"}, {Diamonds, "king"},
	}
	discards2 := ai.ChooseDiscard(hand2)
	if len(discards2) != 0 {
		t.Errorf("Expected 0 discards for Full House, got %v", discards2)
	}
}

func TestChooseDiscard_KeepThreeOfAKind(t *testing.T) {
	ai := DefaultAI
	ai.DiscardSimulations = 1000 // Increased to ensure convergence

	// Case 3: Three Aces, 2 junk.
	// Often best to discard both junk to try for Quads or Full House.
	hand3 := Hand{
		{Hearts, "ace"}, {Spades, "ace"}, {Clubs, "ace"},
		{Hearts, "2"}, {Diamonds, "3"},
	}
	discards3 := ai.ChooseDiscard(hand3)

	valid := false
	// Expecting to discard indices 3 and 4
	if len(discards3) == 2 {
		if (discards3[0] == 3 && discards3[1] == 4) || (discards3[0] == 4 && discards3[1] == 3) {
			valid = true
		}
	}

	if !valid {
		fmt.Printf("Discards: %v\n", discards3)
		// Sometimes strict probability might suggest keeping a kicker if it was high?
		// But 2 and 3 are low.
		t.Errorf("Expected to discard 2 and 3 (indices 3, 4)")
	}
}
