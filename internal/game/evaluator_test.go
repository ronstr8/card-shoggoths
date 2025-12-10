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
	if result != ResultTie {
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
	if result != ResultHand2Wins {
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
	if result != ResultHand2Wins {
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
	if result != ResultTie {
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
	if result != ResultHand1Wins {
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
	if result != ResultHand2Wins {
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
	if result != ResultHand2Wins {
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
	if result != ResultHand1Wins {
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
	if result != ResultHand1Wins {
		t.Errorf("Expected player win with wheel straight over pair, got %d", result)
	}
}

func TestFlushVsFlush(t *testing.T) {
	// Hand 1: Flush Ace high
	hand1 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "J", Suit: "hearts"},
		{Rank: "9", Suit: "hearts"},
		{Rank: "7", Suit: "hearts"},
		{Rank: "5", Suit: "hearts"},
	}

	// Hand 2: Flush King high
	hand2 := []Card{
		{Rank: "K", Suit: "spades"},
		{Rank: "Q", Suit: "spades"},
		{Rank: "J", Suit: "spades"},
		{Rank: "9", Suit: "spades"},
		{Rank: "7", Suit: "spades"},
	}

	result := CompareHands(hand1, hand2)
	if result != ResultHand1Wins {
		t.Errorf("Expected player win with Ace high flush vs King high flush, got %d", result)
	}

	// Same high flush, check kickers
	hand3 := []Card{
		{Rank: "A", Suit: "clubs"},
		{Rank: "J", Suit: "clubs"},
		{Rank: "9", Suit: "clubs"},
		{Rank: "7", Suit: "clubs"},
		{Rank: "4", Suit: "clubs"}, // Lower kicker
	}

	result = CompareHands(hand1, hand3)
	if result != ResultHand1Wins {
		t.Errorf("Expected player win with flush 2nd kicker, got %d", result)
	}
}

func TestHighCardKick(t *testing.T) {
	// Hand 1: A-K-Q-J-9
	hand1 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "K", Suit: "spades"},
		{Rank: "Q", Suit: "diamonds"},
		{Rank: "J", Suit: "clubs"},
		{Rank: "9", Suit: "hearts"},
	}

	// Hand 2: A-K-Q-J-8 (Lower kicker)
	hand2 := []Card{
		{Rank: "A", Suit: "clubs"},
		{Rank: "K", Suit: "hearts"},
		{Rank: "Q", Suit: "spades"},
		{Rank: "J", Suit: "diamonds"},
		{Rank: "8", Suit: "clubs"},
	}

	result := CompareHands(hand1, hand2)
	if result != ResultHand1Wins {
		t.Errorf("Expected player win with better high card kicker, got %d", result)
	}
}

func TestOnePairKick(t *testing.T) {
	// Hand 1: Pair of Aces, K-Q-9
	hand1 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "A", Suit: "spades"},
		{Rank: "K", Suit: "diamonds"},
		{Rank: "Q", Suit: "clubs"},
		{Rank: "9", Suit: "hearts"},
	}

	// Hand 2: Pair of Aces, K-Q-8 (Lower kicker)
	hand2 := []Card{
		{Rank: "A", Suit: "clubs"},
		{Rank: "A", Suit: "diamonds"},
		{Rank: "K", Suit: "hearts"},
		{Rank: "Q", Suit: "spades"},
		{Rank: "8", Suit: "clubs"},
	}

	result := CompareHands(hand1, hand2)
	if result != ResultHand1Wins {
		t.Errorf("Expected player win with pair and better kicker, got %d", result)
	}
}

func TestTwoPairKick(t *testing.T) {
	// Hand 1: AA-KK, Queen kicker
	hand1 := []Card{
		{Rank: "A", Suit: "hearts"},
		{Rank: "A", Suit: "spades"},
		{Rank: "K", Suit: "diamonds"},
		{Rank: "K", Suit: "clubs"},
		{Rank: "Q", Suit: "hearts"},
	}

	// Hand 2: AA-KK, Jack kicker
	hand2 := []Card{
		{Rank: "A", Suit: "clubs"},
		{Rank: "A", Suit: "diamonds"},
		{Rank: "K", Suit: "hearts"},
		{Rank: "K", Suit: "spades"},
		{Rank: "J", Suit: "clubs"},
	}

	result := CompareHands(hand1, hand2)
	if result != ResultHand1Wins {
		t.Errorf("Expected player win with two pair and better kicker, got %d", result)
	}
}

func TestWheelVsHigherStraight(t *testing.T) {
	// Hand 1: 6-high Straight (2-3-4-5-6)
	hand1 := []Card{
		{Rank: "2", Suit: "hearts"},
		{Rank: "3", Suit: "spades"},
		{Rank: "4", Suit: "diamonds"},
		{Rank: "5", Suit: "clubs"},
		{Rank: "6", Suit: "hearts"},
	}

	// Hand 2: Wheel Straight (A-2-3-4-5) - treated as 5-high
	hand2 := []Card{
		{Rank: "A", Suit: "clubs"},
		{Rank: "2", Suit: "hearts"},
		{Rank: "3", Suit: "spades"},
		{Rank: "4", Suit: "diamonds"},
		{Rank: "5", Suit: "clubs"},
	}

	result := CompareHands(hand1, hand2)
	if result != ResultHand1Wins {
		t.Errorf("Expected 6-high straight to beat 5-high (wheel) straight, got %d", result)
	}
}

func TestStraightHighCard(t *testing.T) {
	// Hand 1: 10-high Straight
	hand1 := []Card{
		{Rank: "6", Suit: "hearts"},
		{Rank: "7", Suit: "spades"},
		{Rank: "8", Suit: "diamonds"},
		{Rank: "9", Suit: "clubs"},
		{Rank: "10", Suit: "hearts"},
	}

	// Hand 2: 9-high Straight
	hand2 := []Card{
		{Rank: "5", Suit: "clubs"},
		{Rank: "6", Suit: "hearts"},
		{Rank: "7", Suit: "spades"},
		{Rank: "8", Suit: "diamonds"},
		{Rank: "9", Suit: "clubs"},
	}

	result := CompareHands(hand1, hand2)
	if result != ResultHand1Wins {
		t.Errorf("Expected 10-high straight to beat 9-high straight, got %d", result)
	}
}

func TestFullHouseCompare(t *testing.T) {
	// Hand 1: Queens full of 9s
	hand1 := []Card{
		{Rank: "Q", Suit: "hearts"},
		{Rank: "Q", Suit: "spades"},
		{Rank: "Q", Suit: "diamonds"},
		{Rank: "9", Suit: "clubs"},
		{Rank: "9", Suit: "hearts"},
	}

	// Hand 2: Jacks full of Aces
	hand2 := []Card{
		{Rank: "J", Suit: "clubs"},
		{Rank: "J", Suit: "hearts"},
		{Rank: "J", Suit: "spades"},
		{Rank: "A", Suit: "diamonds"},
		{Rank: "A", Suit: "clubs"},
	}

	result := CompareHands(hand1, hand2)
	if result != ResultHand1Wins {
		t.Errorf("Expected QQQ99 to beat JJJAA, got %d", result)
	}
}

func TestInvalidHandSize(t *testing.T) {
	hand1 := []Card{
		{Rank: "A", Suit: "hearts"},
	}

	val1 := EvaluateHand(hand1)
	if val1.Rank != HighCard {
		t.Errorf("Expected HighCard for invalid hand size")
	}
}
