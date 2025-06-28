package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"card-shoggoths/internal/game"

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

func DiscardHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	sid := getSessionID(w, r)
	state, ok := sessions[sid]
	if !ok {
		state = game.NewGame()
		sessions[sid] = state
	}
	if state.Discarded {
		http.Error(w, "Already discarded", http.StatusBadRequest)
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
	writeJSON(w, state)
}

func ShowdownHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	sid := getSessionID(w, r)
	state, ok := sessions[sid]
	if !ok {
		state = game.NewGame()
		sessions[sid] = state
	}
	state.Showdown = true
	result := game.CompareHands(state.PlayerHand, state.OpponentHand)
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

func getSessionID(w http.ResponseWriter, r *http.Request) string {
    log.Println("Attempting to retrieve session cookie")
	cookie, err := r.Cookie("sid")
	if err == nil {
		return cookie.Value
	}
	id := fmt.Sprintf("%x", time.Now().UnixNano())
	http.SetCookie(w, &http.Cookie{
		Name:  "sid",
		Value: id,
		Path:  "/",
	})
	return id
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
