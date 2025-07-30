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
	sessions[sid] = game.NewGame()
	state := sessions[sid]

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
		Amount int `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !state.PlaceBet(payload.Amount) {
		http.Error(w, "Invalid bet", http.StatusBadRequest)
		return
	}

	// Opponent responds
	response := state.OpponentRespond()

	writeJSON(w, map[string]interface{}{
		"state":   state,
		"message": response,
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

	game.ReplaceCards(&state.Deck, &state.PlayerHand, payload.Indices)
	state.Discarded = true
	state.GamePhase = game.PhaseComplete

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
