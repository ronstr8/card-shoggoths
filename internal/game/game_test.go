package game

import (
	"testing"
)

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

func TestFullGameLoop(t *testing.T) {
	// 1. Initialize Game
	game := NewGame()
	if game.GamePhase != PhaseAnte {
		t.Errorf("Expected PhaseAnte, got %s", game.GamePhase)
	}

	// 2. Collect Ante
	success := game.CollectAnte(10)
	if !success {
		t.Errorf("CollectAnte failed")
	}
	if game.GamePhase != PhasePreDrawBetting {
		t.Errorf("Expected PhasePreDrawBetting, got %s", game.GamePhase)
	}
	if len(game.PlayerHand) != 5 || len(game.OpponentHand) != 5 {
		t.Errorf("Hands not dealt correctly")
	}
	if game.Pot != 20 { // 10 from each
		t.Errorf("Expected Pot 20, got %d", game.Pot)
	}

	// 3. Player Check
	success, _ = game.PlayerAction("check", 0)
	if !success {
		t.Errorf("Player check failed")
	}
	if game.Turn != "opponent" {
		t.Errorf("Expected opponent turn")
	}

	// 4. Opponent Check (AI)
	game.OpponentTurn()

	// Either PhasePreDrawBetting (Opponent Bet) or PhaseDiscard (Opponent Checked)
	if game.GamePhase == PhaseDiscard {
		// Opponent Checked
	} else if game.GamePhase == PhasePreDrawBetting {
		// Opponent Bet
		// Player needs to Call/Fold
		if game.Turn != "player" {
			t.Errorf("Expected player turn to respond to bet")
		}
		// Player Calls
		game.PlayerAction("call", 0) // Amount ignored for call
		if game.GamePhase != PhaseDiscard {
			t.Errorf("Expected PhaseDiscard after call, got %s", game.GamePhase)
		}
	}

	// 5. Discard Phase
	// Player Discards
	game.PerformDiscard([]int{0, 1}) // Discard first 2

	// PerformDiscard calls Opponent discard and NextPhase internally
	if game.GamePhase != PhasePostDrawBetting {
		t.Errorf("Expected PhasePostDrawBetting, got %s", game.GamePhase)
	}

	// 6. Post-Draw Betting
	// Player Checks
	game.PlayerAction("check", 0)

	// Opponent Turn
	game.OpponentTurn()

	if game.GamePhase == PhasePostDrawBetting {
		// Opponent Bet
		game.PlayerAction("call", 0)
	}

	if game.GamePhase != PhaseComplete {
		t.Errorf("Expected PhaseComplete, got %s", game.GamePhase)
	}

	if game.Winner == "" {
		t.Errorf("Expected a winner")
	}
}
