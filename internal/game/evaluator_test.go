package game

import (
	"testing"
)

func mustHand(cards ...Card) Hand {
	return cards
}

func TestCompareHandsTie(t *testing.T) {
	h := Hand{
		{Suit: Spades, Rank: "2"},
		{Suit: Hearts, Rank: "5"},
		{Suit: Diamonds, Rank: "9"},
		{Suit: Clubs, Rank: "J"},
		{Suit: Hearts, Rank: "A"},
	}
	result := CompareHands(h, h)
	if result != "It's a tie — both with High Card!" {
		t.Errorf("Expected tie, got %s", result)
	}
}

func TestHighCardVsOnePair(t *testing.T) {
	player := mustHand(
		Card{Suit: Spades, Rank: "2"},
		Card{Suit: Hearts, Rank: "5"},
		Card{Suit: Diamonds, Rank: "9"},
		Card{Suit: Clubs, Rank: "J"},
		Card{Suit: Hearts, Rank: "A"},
	)
	opponent := mustHand(
		Card{Suit: Spades, Rank: "4"},
		Card{Suit: Hearts, Rank: "4"},
		Card{Suit: Diamonds, Rank: "8"},
		Card{Suit: Clubs, Rank: "Q"},
		Card{Suit: Hearts, Rank: "K"},
	)
	winner := CompareHands(player, opponent)
	if winner != "Opponent wins with a One Pair beating a High Card!" {
		t.Errorf("Expected opponent win with pair over high card, got %s", winner)
	}
}

func TestFullHouseBeatsFlush(t *testing.T) {
	flush := mustHand(
		Card{Suit: Hearts, Rank: "2"},
		Card{Suit: Hearts, Rank: "5"},
		Card{Suit: Hearts, Rank: "8"},
		Card{Suit: Hearts, Rank: "J"},
		Card{Suit: Hearts, Rank: "Q"},
	)
	fullhouse := mustHand(
		Card{Suit: Clubs, Rank: "K"},
		Card{Suit: Spades, Rank: "K"},
		Card{Suit: Hearts, Rank: "K"},
		Card{Suit: Diamonds, Rank: "10"},
		Card{Suit: Clubs, Rank: "10"},
	)
	winner := CompareHands(flush, fullhouse)
	if winner != "Opponent wins with a Full House beating a Flush!" {
		t.Errorf("Expected opponent win with Full House, got %s", winner)
	}
}

func TestTieOnExactSameCards(t *testing.T) {
	hand := mustHand(
		Card{Suit: Clubs, Rank: "2"},
		Card{Suit: Spades, Rank: "5"},
		Card{Suit: Hearts, Rank: "9"},
		Card{Suit: Diamonds, Rank: "J"},
		Card{Suit: Clubs, Rank: "A"},
	)
	result := CompareHands(hand, hand)
	if result != "It's a tie — both with High Card!" {
		t.Errorf("Expected tie with both having high cards, got %s", result)
	}
}

func TestStraightBeatsThreeOfAKind(t *testing.T) {
	straight := mustHand(
		Card{Suit: Clubs, Rank: "5"},
		Card{Suit: Diamonds, Rank: "6"},
		Card{Suit: Hearts, Rank: "7"},
		Card{Suit: Spades, Rank: "8"},
		Card{Suit: Clubs, Rank: "9"},
	)
	three := mustHand(
		Card{Suit: Hearts, Rank: "Q"},
		Card{Suit: Spades, Rank: "Q"},
		Card{Suit: Clubs, Rank: "Q"},
		Card{Suit: Diamonds, Rank: "3"},
		Card{Suit: Hearts, Rank: "2"},
	)
	winner := CompareHands(straight, three)
	if winner != "Opponent wins with a Three of a Kind beating a Straight!" {
		t.Errorf("Expected opponent win with three of a kind over straight, got %s", winner)
	}
}
