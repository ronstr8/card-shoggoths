package game

import (
	"sort"
)

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
)

type RankedHand struct {
	Rank   HandRank
	Values []int
}

// map ranks to their values
var rankValueMap = map[Rank]int{
	"2": 2, "3": 3, "4": 4, "5": 5, "6": 6, "7": 7,
	"8": 8, "9": 9, "10": 10, "J": 11, "Q": 12, "K": 13, "A": 14,
}

func getCounts(hand Hand) (map[int]int, []int) {
	counts := map[int]int{}
	values := []int{}
	for _, card := range hand {
		val := rankValueMap[card.Rank]
		counts[val]++
		values = append(values, val)
	}
	return counts, values
}

func isFlush(hand Hand) bool {
	firstSuit := hand[0].Suit
	for _, card := range hand[1:] {
		if card.Suit != firstSuit {
			return false
		}
	}
	return true
}

func isStraight(values []int) bool {
	sort.Ints(values)
	for i := 1; i < len(values); i++ {
		if values[i] != values[i-1]+1 {
			// check for wheel (A-2-3-4-5)
			if i == 4 && values[0] == 2 && values[4] == 14 {
				return true
			}
			return false
		}
	}
	return true
}

func sortedValueGroups(counts map[int]int) []int {
	type pair struct {
		val   int
		count int
	}
	var pairs []pair
	for val, count := range counts {
		pairs = append(pairs, pair{val, count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].val > pairs[j].val
		}
		return pairs[i].count > pairs[j].count
	})
	var sorted []int
	for _, p := range pairs {
		for i := 0; i < p.count; i++ {
			sorted = append(sorted, p.val)
		}
	}
	return sorted
}

func RankHand(hand Hand) RankedHand {
	counts, values := getCounts(hand)
	flush := isFlush(hand)
	straight := isStraight(values)

	switch {
	case flush && straight:
		return RankedHand{Rank: StraightFlush, Values: sortedValueGroups(counts)}
	case hasCount(counts, 4):
		return RankedHand{Rank: FourOfAKind, Values: sortedValueGroups(counts)}
	case hasCount(counts, 3) && hasCount(counts, 2):
		return RankedHand{Rank: FullHouse, Values: sortedValueGroups(counts)}
	case flush:
		return RankedHand{Rank: Flush, Values: sortedValueGroups(counts)}
	case straight:
		return RankedHand{Rank: Straight, Values: sortedValueGroups(counts)}
	case hasCount(counts, 3):
		return RankedHand{Rank: ThreeOfAKind, Values: sortedValueGroups(counts)}
	case numPairs(counts) == 2:
		return RankedHand{Rank: TwoPair, Values: sortedValueGroups(counts)}
	case hasCount(counts, 2):
		return RankedHand{Rank: OnePair, Values: sortedValueGroups(counts)}
	default:
		return RankedHand{Rank: HighCard, Values: sortedValueGroups(counts)}
	}
}

func hasCount(counts map[int]int, n int) bool {
	for _, count := range counts {
		if count == n {
			return true
		}
	}
	return false
}

func numPairs(counts map[int]int) int {
	n := 0
	for _, count := range counts {
		if count == 2 {
			n++
		}
	}
	return n
}

func CompareHands(player, opponent Hand) string {
	r1 := RankHand(player)
	r2 := RankHand(opponent)

	if r1.Rank > r2.Rank {
		return "Player wins!"
	} else if r2.Rank > r1.Rank {
		return "Lucky Pete wins!"
	} else {
		for i := range r1.Values {
			if r1.Values[i] > r2.Values[i] {
				return "Player wins!"
			} else if r2.Values[i] > r1.Values[i] {
				return "Lucky Pete wins!"
			}
		}
		return "It's a tie!"
	}
}
