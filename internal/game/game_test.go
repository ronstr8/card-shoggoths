package game

import "testing"

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 52 {
		t.Fatalf("expected 52 cards, got %d", len(deck))
	}
}

func TestDealHand(t *testing.T) {
	deck := NewDeck()
	hand := DealHand(&deck, 5)
	if len(hand) != 5 {
		t.Errorf("expected 5 cards, got %d", len(hand))
	}
}
