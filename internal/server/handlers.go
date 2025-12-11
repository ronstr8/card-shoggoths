package server

import (
	"card-shoggoths/internal/game"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// Session-based state tracking
var sessions = map[string]*game.GameState{}
var mu sync.Mutex

func DealHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	sid := getSessionID(w, r)
	if _, ok := sessions[sid]; !ok {
		sessions[sid] = game.NewGame()
	}
	state := sessions[sid]

	// If the game is complete (or just starting), start a new round
	if state.GamePhase == game.PhaseComplete || state.GamePhase == game.PhaseAnte {
		state.NewRound()
	} else {
		// If dealing in middle of game, maybe reset?
		// For safety, let's treat "Deal" as "New Hand/Reset"
		state.NewRound()
	}

	// Auto-collect ante to start the game properly
	state.CollectAnte(10)

	writeJSON(w, state)
}

func BetHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	sid := getSessionID(w, r)
	state, ok := sessions[sid]
	if !ok {
		http.Error(w, "No active game", http.StatusBadRequest)
		return
	}

	var payload struct {
		Action string `json:"action"` // explicit action
		Amount int    `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Infer action from amount if not provided
	toCall := state.CurrentBet - state.PlayerBet
	action := payload.Action
	actionAmount := 0

	if action == "" {
		// Legacy inference logic
		if payload.Amount == toCall {
			action = "call"
		} else if payload.Amount > toCall {
			if toCall == 0 && state.CurrentBet == 0 {
				action = "bet"
				actionAmount = payload.Amount
			} else {
				action = "raise"
				actionAmount = payload.Amount - toCall // Raise BY X
			}
		} else if payload.Amount == 0 && toCall == 0 {
			action = "check"
		} else {
			http.Error(w, "Invalid bet amount", http.StatusBadRequest)
			return
		}
	} else {
		// Handle explicit actions
		if action == "bet" || action == "raise" {
			// For raise/bet, we need to handle the amount correctly
			// If action is raise, assume payload.Amount is the RAISE BY amount for consistency with frontend
			// Or should we support absolute? stick to RAISE BY for now as per previous logic
			if action == "raise" {
				// Raise logic expects 'amount' to be the delta?
				// Previous inference: actionAmount = payload.Amount - toCall
				// But if frontend sends explicit "raise" + amount... let's assume `amount` IS the raise-by amount.
				actionAmount = payload.Amount
			} else if action == "bet" {
				actionAmount = payload.Amount
			}
		}
	}

	success, msg := state.PlayerAction(action, actionAmount)
	if !success {
		errMsg := msg
		if errMsg == "" {
			errMsg = "Invalid action"
		}
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	// Trigger opponent turn if it's their turn
	if state.Turn == "opponent" {
		state.OpponentTurn()
	}

	writeJSON(w, map[string]interface{}{
		"state":   state,
		"message": state.LastAction,
	})
}

func DiscardHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	sid := getSessionID(w, r)
	state, ok := sessions[sid]
	if !ok {
		state = game.NewGame()
		sessions[sid] = state
	}

	if !state.CanDiscard() {
		http.Error(w, "Cannot discard at this time", http.StatusBadRequest)
		return
	}

	var payload struct {
		Indices []int `json:"indices"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	state.PerformDiscard(payload.Indices)

	writeJSON(w, state)
}

func ShowdownHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	sid := getSessionID(w, r)
	state, ok := sessions[sid]
	if !ok {
		http.Error(w, "No active game", http.StatusBadRequest)
		return
	}

	if !state.CanShowdown() {
		http.Error(w, "Cannot showdown at this time", http.StatusBadRequest)
		return
	}

	state.CompleteShowdown()
	result := game.CompareHandsForDisplay(state.PlayerHand, state.OpponentHand)

	writeJSON(w, map[string]interface{}{
		"result": result,
		"state":  state,
	})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Println("JSON encode error:", err)
	}
}

// getSessionID will return an existing session_id cookie,
// or create a new one and hand it back.
func getSessionID(w http.ResponseWriter, r *http.Request) string {
	log.Println("Attempting to retrieve session cookie")
	// First, try to read the "session_id" cookie
	if cookie, err := r.Cookie("session_id"); err == nil {
		return cookie.Value
	}
	// If missing, make a new UUID, set a cookie, and return it
	sessionID := uuid.NewString()
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		// you can also set Secure, SameSite, etc. here
	})
	return sessionID
}

func createNewSession(w http.ResponseWriter, r *http.Request) *game.GameState {
	sessionID := uuid.NewString()
	gs := game.NewGame()
	mu.Lock()
	sessions[sessionID] = gs
	mu.Unlock()
	cookie := &http.Cookie{
		Name:  "session_id",
		Value: sessionID,
		Path:  "/",
	}
	http.SetCookie(w, cookie)
	log.Printf("New session created with ID: %s", sessionID)
	return gs
}

func ClearSessionHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "No session cookie found", http.StatusBadRequest)
		return
	}
	mu.Lock()
	defer mu.Unlock()
	delete(sessions, cookie.Value)
	log.Printf("Cleared session for ID: %s", cookie.Value)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Session cleared"))
}
