package server

import (
	"card-shoggoths/internal/game"
	"card-shoggoths/internal/store"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
)

var gameStore store.GameStore

// Init sets the storage backend
func Init(s store.GameStore) {
	gameStore = s
}
func getSessionID(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie("session_id"); err == nil {
		log.Printf("[DEBUG] Cookie found: %s", cookie.Value)
		return cookie.Value
	}
	sessionID := uuid.NewString()
	log.Printf("[DEBUG] No cookie found. Generated new Session ID: %s", sessionID)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return sessionID
}

func getGame(w http.ResponseWriter, r *http.Request) (*game.GameState, string) {
	sid := getSessionID(w, r)
	g, err := gameStore.Load(sid)
	if err != nil {
		log.Printf("[DEBUG] Load failed for session %s: %v", sid, err)
		return nil, sid
	}
	log.Printf("[DEBUG] Loaded game for session %s", sid)
	return g, sid
}

func saveGame(id string, g *game.GameState) error {
	if err := gameStore.Save(id, g); err != nil {
		log.Printf("[ERROR] Failed to save session %s: %v", id, err)
		return err
	}
	log.Printf("[DEBUG] Saved game for session %s", id)
	return nil
}

func DealHandler(w http.ResponseWriter, r *http.Request) {
	g, sid := getGame(w, r)
	log.Printf("[DEBUG] DealHandler: Session %s", sid)

	if g == nil {
		g = game.NewGame()
		g.ID = sid
		log.Printf("[DEBUG] Created new game object for session %s", sid)
	} else {
		g.NewRound()
	}

	g.CollectAnte(10)
	if err := saveGame(sid, g); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save game state: %v", err), http.StatusInternalServerError)
		return
	}

	// Ancient One comments on the deal
	go SendAncientMessage(sid, "deal")

	writeJSON(w, g)
}

func StateHandler(w http.ResponseWriter, r *http.Request) {
	g, _ := getGame(w, r)
	if g == nil {
		// No game exists yet, return empty/null
		writeJSON(w, nil)
		return
	}
	writeJSON(w, g)
}

func ActionHandler(w http.ResponseWriter, r *http.Request) {
	g, sid := getGame(w, r)
	if g == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	var payload struct {
		Action string `json:"action"`
		Amount int    `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Auto-fill defaults if legacy
	if payload.Action == "" {
		// Basic inference or default to 'call'/'check' logic?
		// Prefer explicit actions now.
		http.Error(w, "Action required", http.StatusBadRequest)
		return
	}

	success, msg := g.PlayerAction(payload.Action, payload.Amount)
	if !success {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	g.OpponentTurn()
	if err := saveGame(sid, g); err != nil {
		// Log but maybe don't fail the request completely since action succeeded?
		// Actually, if we don't save, state is lost. Better to warn.
		http.Error(w, "State save failed", http.StatusInternalServerError)
		return
	}

	// Ancient One reacts to player action
	switch payload.Action {
	case "fold":
		go SendAncientMessage(sid, "player_fold")
	case "bet", "call", "raise", "check":
		go SendAncientMessage(sid, "player_bet")
	}

	writeJSON(w, g)
}

func DiscardHandler(w http.ResponseWriter, r *http.Request) {
	g, sid := getGame(w, r)
	if g == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	if !g.CanDiscard() {
		http.Error(w, "Cannot discard now", http.StatusBadRequest)
		return
	}

	var payload struct {
		Indices []int `json:"indices"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	g.PerformDiscard(payload.Indices)
	if err := saveGame(sid, g); err != nil {
		http.Error(w, "State save failed", http.StatusInternalServerError)
		return
	}
	writeJSON(w, g)
}

func ShowdownHandler(w http.ResponseWriter, r *http.Request) {
	g, sid := getGame(w, r)
	if g == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	if !g.CanShowdown() {
		http.Error(w, "Cannot showdown now", http.StatusBadRequest)
		return
	}

	g.CompleteShowdown()
	saveGame(sid, g)

	// Ancient One reacts to outcome
	if g.Winner == g.Players[0].Name {
		go SendAncientMessage(sid, "player_wins")
	} else if g.Winner == g.Players[1].Name {
		go SendAncientMessage(sid, "ancient_wins")
	}

	// Frontend expects { result: ..., state: ... }
	playerHand := g.RoundStates[0].Hand
	opponentHand := g.RoundStates[1].Hand
	result := game.CompareHandsForDisplay(playerHand, opponentHand)

	writeJSON(w, map[string]interface{}{
		"result": result,
		"state":  g,
	})
}

func ClearSessionHandler(w http.ResponseWriter, r *http.Request) {
	// Not easily supported with SQLite without ID.
	// For now just ignore or implement delete if strictly needed.
	w.WriteHeader(http.StatusOK)
}

func RebuyHandler(w http.ResponseWriter, r *http.Request) {
	g, sid := getGame(w, r)
	if g == nil {
		// No game exists, create one
		g = game.NewGame()
		g.ID = sid
		g.CollectAnte(10) // Auto-start
	} else {
		// Reset logic: New Game completely
		newGame := game.NewGame()
		newGame.ID = sid
		g = newGame
		g.CollectAnte(10)
	}
	saveGame(sid, g)
	writeJSON(w, g)
}

func ESPStartHandler(w http.ResponseWriter, r *http.Request) {
	g, sid := getGame(w, r)
	if g == nil {
		http.Error(w, "Game not found. Deal first.", http.StatusNotFound)
		return
	}

	ok, msg := g.StartESP()
	if !ok {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	saveGame(sid, g)
	writeJSON(w, g)
}

func ESPGuessHandler(w http.ResponseWriter, r *http.Request) {
	g, sid := getGame(w, r)
	if g == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	var payload struct {
		Index1 int `json:"index1"`
		Index2 int `json:"index2"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	correct, _ := g.GuessESP(payload.Index1, payload.Index2)
	saveGame(sid, g)

	writeJSON(w, map[string]interface{}{
		"correct": correct,
		"state":   g,
	})
}

func ESPExitHandler(w http.ResponseWriter, r *http.Request) {
	g, sid := getGame(w, r)
	if g == nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	g.ExitESP()
	saveGame(sid, g)
	writeJSON(w, g)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
