package game

import "testing"

func TestCompareHandsTie(t *testing.T) {
	// Both hands have the same high card
	hand1 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "K", Suit: "spades"},
		{Rank: "Q", Suit: "diamonds"},
		{Rank: "J", Suit: "clubs"},
		{Rank: "9", Suit: "hearts"},
	}
	hand2 := []Card{
		{Rank: "A", Suit: "clubs"},
		{Rank: "K", Suit: "hearts"},
		{Rank: "Q", Suit: "spades"},
		{Rank: "J", Suit: "diamonds"},
		{Rank: "9", Suit: "clubs"},
	}
	
	result := CompareHands(hand1, hand2)
	if result != 0 {
		t.Errorf("Expected tie, got %d", result)
	}
}

func TestHighCardVsOnePair(t *testing.T) {
	// Hand 1: High card A-K-Q-J-9
	hand1 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "K", Suit: "spades"},
		{Rank: "Q", Suit: "diamonds"},
		{Rank: "J", Suit: "clubs"},
		{Rank: "9", Suit: "hearts"},
	}
	
	// Hand 2: Pair of 2s
	hand2 := []Card{
		{Rank: "2", Suit: "hearts"},
		{Rank: "2", Suit: "spades"},
		{Rank: "7", Suit: "diamonds"},
		{Rank: "8", Suit: "clubs"},
		{Rank: "9", Suit: "hearts"},
	}
	
	result := CompareHands(hand1, hand2)
	if result != -1 {
		t.Errorf("Expected opponent win with pair over high card, got %d", result)
	}
}

func TestFullHouseBeatsFlush(t *testing.T) {
	// Hand 1: Flush in hearts
	hand1 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "K", Suit: "hearts"},
		{Rank: "Q", Suit: "hearts"},
		{Rank: "J", Suit: "hearts"},
		{Rank: "9", Suit: "hearts"},
	}
	
	// Hand 2: Full house (3 Kings, 2 Twos)
	hand2 := []Card{
		{Rank: "K", Suit: "hearts"},
		{Rank: "K", Suit: "spades"},
		{Rank: "K", Suit: "diamonds"},
		{Rank: "2", Suit: "clubs"},
		{Rank: "2", Suit: "hearts"},
	}
	
	result := CompareHands(hand1, hand2)
	if result != -1 {
		t.Errorf("Expected opponent win with Full House, got %d", result)
	}
}

func TestTieOnExactSameCards(t *testing.T) {
	// Both hands have identical high cards
	hand1 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "K", Suit: "spades"},
		{Rank: "Q", Suit: "diamonds"},
		{Rank: "J", Suit: "clubs"},
		{Rank: "9", Suit: "hearts"},
	}
	hand2 := []Card{
		{Rank: "A", Suit: "clubs"},
		{Rank: "K", Suit: "hearts"},
		{Rank: "Q", Suit: "spades"},
		{Rank: "J", Suit: "diamonds"},
		{Rank: "9", Suit: "clubs"},
	}
	
	result := CompareHands(hand1, hand2)
	if result != 0 {
		t.Errorf("Expected tie with both having high cards, got %d", result)
	}
}

func TestStraightBeatsThreeOfAKind(t *testing.T) {
	// Hand 1: Straight (5-6-7-8-9)
	hand1 := []Card{
		{Rank: "5", Suit: "hearts"},
		{Rank: "6", Suit: "spades"},
		{Rank: "7", Suit: "diamonds"},
		{Rank: "8", Suit: "clubs"},
		{Rank: "9", Suit: "hearts"},
	}
	
	// Hand 2: Three of a kind (3 Aces)
	hand2 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "A", Suit: "spades"},
		{Rank: "A", Suit: "diamonds"},
		{Rank: "K", Suit: "clubs"},
		{Rank: "Q", Suit: "hearts"},
	}
	
	result := CompareHands(hand1, hand2)
	if result != 1 {
		t.Errorf("Expected player win with straight over three of a kind, got %d", result)
	}
}

func TestThreeOfAKindBeatsTwoPair(t *testing.T) {
	// Hand 1: Two pair (Kings and Queens)
	hand1 := []Card{
		{Rank: "K", Suit: "hearts"},
		{Rank: "K", Suit: "spades"},
		{Rank: "Q", Suit: "diamonds"},
		{Rank: "Q", Suit: "clubs"},
		{Rank: "9", Suit: "hearts"},
	}
	
	// Hand 2: Three of a kind (3 Twos)
	hand2 := []Card{
		{Rank: "2", Suit: "hearts"},
		{Rank: "2", Suit: "spades"},
		{Rank: "2", Suit: "diamonds"},
		{Rank: "A", Suit: "clubs"},
		{Rank: "K", Suit: "hearts"},
	}
	
	result := CompareHands(hand1, hand2)
	if result != -1 {
		t.Errorf("Expected opponent win with three of a kind over two pair, got %d", result)
	}
}

func TestFlushBeatsStraight(t *testing.T) {
	// Hand 1: Straight (5-6-7-8-9)
	hand1 := []Card{
		{Rank: "5", Suit: "hearts"},
		{Rank: "6", Suit: "spades"},
		{Rank: "7", Suit: "diamonds"},
		{Rank: "8", Suit: "clubs"},
		{Rank: "9", Suit: "hearts"},
	}
	
	// Hand 2: Flush (all hearts)
	hand2 := []Card{
		{Rank: "2", Suit: "hearts"},
		{Rank: "4", Suit: "hearts"},
		{Rank: "6", Suit: "hearts"},
		{Rank: "8", Suit: "hearts"},
		{Rank: "J", Suit: "hearts"},
	}
	
	result := CompareHands(hand1, hand2)
	if result != -1 {
		t.Errorf("Expected opponent win with flush over straight, got %d", result)
	}
}

func TestRoyalFlush(t *testing.T) {
	// Hand 1: Royal flush in hearts
	hand1 := []Card{
		{Rank: "10", Suit: "hearts"},
		{Rank: "J", Suit: "hearts"},
		{Rank: "Q", Suit: "hearts"},
		{Rank: "K", Suit: "hearts"},
		{Rank: "A", Suit: "hearts"},
	}
	
	// Hand 2: Four of a kind (4 Aces)
	hand2 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "A", Suit: "spades"},
		{Rank: "A", Suit: "diamonds"},
		{Rank: "A", Suit: "clubs"},
		{Rank: "K", Suit: "hearts"},
	}
	
	result := CompareHands(hand1, hand2)
	if result != 1 {
		t.Errorf("Expected player win with royal flush over four of a kind, got %d", result)
	}
}

func TestWheelStraight(t *testing.T) {
	// Hand 1: Wheel straight (A-2-3-4-5)
	hand1 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "2", Suit: "spades"},
		{Rank: "3", Suit: "diamonds"},
		{Rank: "4", Suit: "clubs"},
		{Rank: "5", Suit: "hearts"},
	}
	
	// Hand 2: Pair of Kings
	hand2 := []Card{
		{Rank: "K", Suit: "hearts"},
		{Rank: "K", Suit: "spades"},
		{Rank: "Q", Suit: "diamonds"},
		{Rank: "J", Suit: "clubs"},
		{Rank: "10", Suit: "hearts"},
	}
	
	result := CompareHands(hand1, hand2)
	if result != 1 {
		t.Errorf("Expected player win with wheel straight over pair, got %d", result)
	}
}
