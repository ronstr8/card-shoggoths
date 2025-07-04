package game

import (
	"fmt"
	"sort"
	"strconv"
)

// HandRank represents the ranking of a poker hand
type HandRank int

const (
	HighCard HandRank = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

// HandValue represents the value of a poker hand for comparison
type HandValue struct {
	Rank     HandRank
	Primary  int // Primary ranking value (e.g., the rank of the pair)
	Secondary int // Secondary ranking value (e.g., kicker)
	Kickers  []int // Additional kickers for tie-breaking
}

// CardValue returns the numeric value of a card rank
func CardValue(rank Rank) int {
	switch string(rank) {
	case "A":
		return 14
	case "K":
		return 13
	case "Q":
		return 12
	case "J":
		return 11
	default:
		if val, err := strconv.Atoi(string(rank)); err == nil {
			return val
		}
		return 0
	}
}

// EvaluateHand evaluates a poker hand and returns its HandValue
func EvaluateHand(hand []Card) HandValue {
	if len(hand) != 5 {
		return HandValue{Rank: HighCard}
	}

	// Convert cards to values and suits
	values := make([]int, len(hand))
	suits := make([]string, len(hand))
	for i, card := range hand {
		values[i] = CardValue(card.Rank)
		suits[i] = string(card.Suit)
	}

	// Sort values for easier analysis
	sort.Ints(values)

	// Count occurrences of each value
	counts := make(map[int]int)
	for _, val := range values {
		counts[val]++
	}

	// Check for flush
	isFlush := true
	for i := 1; i < len(suits); i++ {
		if suits[i] != suits[0] {
			isFlush = false
			break
		}
	}

	// Check for straight
	isStraight := true
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			isStraight = false
			break
		}
	}

	// Special case: A-2-3-4-5 straight (wheel)
	if !isStraight && len(values) == 5 {
		if values[0] == 2 && values[1] == 3 && values[2] == 4 && values[3] == 5 && values[4] == 14 {
			isStraight = true
			values = []int{1, 2, 3, 4, 5} // Treat ace as 1 for wheel
		}
	}

	// Determine hand rank
	var countGroups []int
	for _, count := range counts {
		countGroups = append(countGroups, count)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(countGroups)))

	// Royal flush check
	if isFlush && isStraight && values[4] == 14 && values[0] == 10 {
		return HandValue{Rank: RoyalFlush, Primary: 14}
	}

	// Straight flush
	if isFlush && isStraight {
		return HandValue{Rank: StraightFlush, Primary: values[4]}
	}

	// Four of a kind
	if len(countGroups) == 2 && countGroups[0] == 4 {
		var fourKind, kicker int
		for val, count := range counts {
			if count == 4 {
				fourKind = val
			} else {
				kicker = val
			}
		}
		return HandValue{Rank: FourOfAKind, Primary: fourKind, Secondary: kicker}
	}

	// Full house
	if len(countGroups) == 2 && countGroups[0] == 3 && countGroups[1] == 2 {
		var threeKind, pair int
		for val, count := range counts {
			if count == 3 {
				threeKind = val
			} else {
				pair = val
			}
		}
		return HandValue{Rank: FullHouse, Primary: threeKind, Secondary: pair}
	}

	// Flush
	if isFlush {
		return HandValue{Rank: Flush, Kickers: reverseSlice(values)}
	}

	// Straight
	if isStraight {
		return HandValue{Rank: Straight, Primary: values[4]}
	}

	// Three of a kind
	if len(countGroups) == 3 && countGroups[0] == 3 {
		var threeKind int
		var kickers []int
		for val, count := range counts {
			if count == 3 {
				threeKind = val
			} else {
				kickers = append(kickers, val)
			}
		}
		sort.Sort(sort.Reverse(sort.IntSlice(kickers)))
		return HandValue{Rank: ThreeOfAKind, Primary: threeKind, Kickers: kickers}
	}

	// Two pair
	if len(countGroups) == 3 && countGroups[0] == 2 && countGroups[1] == 2 {
		var pairs []int
		var kicker int
		for val, count := range counts {
			if count == 2 {
				pairs = append(pairs, val)
			} else {
				kicker = val
			}
		}
		sort.Sort(sort.Reverse(sort.IntSlice(pairs)))
		return HandValue{Rank: TwoPair, Primary: pairs[0], Secondary: pairs[1], Kickers: []int{kicker}}
	}

	// One pair
	if len(countGroups) == 4 && countGroups[0] == 2 {
		var pair int
		var kickers []int
		for val, count := range counts {
			if count == 2 {
				pair = val
			} else {
				kickers = append(kickers, val)
			}
		}
		sort.Sort(sort.Reverse(sort.IntSlice(kickers)))
		return HandValue{Rank: OnePair, Primary: pair, Kickers: kickers}
	}

	// High card
	return HandValue{Rank: HighCard, Kickers: reverseSlice(values)}
}

// CompareHands compares two poker hands and returns:
// 1 if hand1 wins, -1 if hand2 wins, 0 if tie
func CompareHands(hand1, hand2 []Card) int {
	val1 := EvaluateHand(hand1)
	val2 := EvaluateHand(hand2)

	// Compare hand ranks first
	if val1.Rank > val2.Rank {
		return 1
	} else if val1.Rank < val2.Rank {
		return -1
	}

	// Same rank, compare primary values
	if val1.Primary > val2.Primary {
		return 1
	} else if val1.Primary < val2.Primary {
		return -1
	}

	// Same primary, compare secondary values
	if val1.Secondary > val2.Secondary {
		return 1
	} else if val1.Secondary < val2.Secondary {
		return -1
	}

	// Compare kickers
	for i := 0; i < len(val1.Kickers) && i < len(val2.Kickers); i++ {
		if val1.Kickers[i] > val2.Kickers[i] {
			return 1
		} else if val1.Kickers[i] < val2.Kickers[i] {
			return -1
		}
	}

	// If all comparisons are equal, it's a tie
	return 0
}

// Helper function to reverse a slice
func reverseSlice(s []int) []int {
	result := make([]int, len(s))
	for i, v := range s {
		result[len(s)-1-i] = v
	}
	return result
}

// GetHandName returns a human-readable name for the hand rank
func GetHandName(rank HandRank) string {
	switch rank {
	case HighCard:
		return "High Card"
	case OnePair:
		return "One Pair"
	case TwoPair:
		return "Two Pair"
	case ThreeOfAKind:
		return "Three of a Kind"
	case Straight:
		return "Straight"
	case Flush:
		return "Flush"
	case FullHouse:
		return "Full House"
	case FourOfAKind:
		return "Four of a Kind"
	case StraightFlush:
		return "Straight Flush"
	case RoyalFlush:
		return "Royal Flush"
	default:
		return "Unknown"
	}
}

// CompareHandsString compares two poker hands and returns a string result
// for compatibility with existing game logic
func CompareHandsString(hand1, hand2 []Card) string {
	result := CompareHands(hand1, hand2)
	if result > 0 {
		return "player"
	} else if result < 0 {
		return "opponent"
	} else {
		return "tie"
	}
}

// HandComparisonResult holds detailed comparison information for display
type HandComparisonResult struct {
	Winner          string
	PlayerHandName  string
	OpponentHandName string
	Message         string
}

// CompareHandsForDisplay compares two poker hands and returns detailed result information
// for display to the user
func CompareHandsForDisplay(playerHand, opponentHand []Card) HandComparisonResult {
	playerValue := EvaluateHand(playerHand)
	opponentValue := EvaluateHand(opponentHand)
	
	result := CompareHands(playerHand, opponentHand)
	
	playerHandName := GetHandName(playerValue.Rank)
	opponentHandName := GetHandName(opponentValue.Rank)
	
	var winner string
	var message string
	
	if result > 0 {
		winner = "player"
		message = fmt.Sprintf("You win with %s! The Ancient One had %s.", playerHandName, opponentHandName)
	} else if result < 0 {
		winner = "opponent"
		message = fmt.Sprintf("The Ancient One wins with %s! You had %s.", opponentHandName, playerHandName)
	} else {
		winner = "tie"
		message = fmt.Sprintf("It's a tie! Both hands have %s.", playerHandName)
	}
	
	return HandComparisonResult{
		Winner:          winner,
		PlayerHandName:  playerHandName,
		OpponentHandName: opponentHandName,
		Message:         message,
	}
}
