package server

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// ChatMessage represents a message in the chat
type ChatMessage struct {
	Sender    string `json:"sender"` // "ancient_one" or "player"
	Text      string `json:"text"`
	Type      string `json:"type"` // "speech", "emote", "system"
	Timestamp int64  `json:"timestamp"`
}

// ClientConnection wraps a websocket connection
type ClientConnection struct {
	Conn      *websocket.Conn
	SessionID string
	mu        sync.Mutex
}

var (
	clients   = make(map[string]*ClientConnection)
	clientsMu sync.RWMutex
)

// Ancient One's pre-defined messages by situation
var ancientQuips = map[string][]string{
	"deal": {
		"*shuffles cards with tentacles*",
		"The cards whisper secrets...",
		"Fate is dealt anew.",
		"Let us see what the void reveals.",
		"*eyes glow faintly*",
	},
	"player_bet": {
		"Bold... or foolish?",
		"You wager your sanity freely.",
		"*chuckles in frequencies below human hearing*",
		"The stakes... rise.",
		"Interesting.",
	},
	"player_fold": {
		"Wisdom... or cowardice?",
		"The void notes your retreat.",
		"*nods slowly*",
		"Self-preservation. How... mortal.",
		"You yield to the inevitable.",
	},
	"ancient_wins": {
		"Your sanity feeds me.",
		"*absorbs essence*",
		"The cosmos favors the eternal.",
		"Another fragment of your mind... mine.",
		"Delicious despair.",
	},
	"player_wins": {
		"*hisses* Impossible...",
		"A temporary setback.",
		"You have... luck. For now.",
		"*tentacles twitch with agitation*",
		"The void is patient.",
	},
	"idle": {
		"*clears throat in a tone that predates language*",
		"Time moves differently for immortals... but still, it moves.",
		"*taps table with appendage*",
		"Do mortals always deliberate this long?",
		"The cards grow cold waiting.",
		"*stares into your soul*",
		"Eternity stretches before us... but perhaps not THAT long.",
	},
	"esp_start": {
		"Ah, you dare peer beyond the veil?",
		"*opens third eye*",
		"The patterns of reality shimmer...",
		"Focus... if your feeble mind can.",
	},
	"esp_success": {
		"*surprised gurgle* You... saw?",
		"The gift stirs within you.",
		"*grudging respect*",
	},
	"esp_fail": {
		"*laughs in cosmic horror*",
		"Your third eye remains... clouded.",
		"The visions elude you.",
	},
	"greeting": {
		"Welcome, mortal. Sit. Play. Lose your mind.",
		"Another soul seeks to challenge the void.",
		"*manifests at the table* Shall we begin?",
	},
}

// GetAncientQuip returns a random quip for a situation
func GetAncientQuip(situation string) string {
	quips, ok := ancientQuips[situation]
	if !ok || len(quips) == 0 {
		return "*stares inscrutably*"
	}
	return quips[rand.Intn(len(quips))]
}

// SendToClient sends a chat message to a specific client
func SendToClient(sessionID string, msg ChatMessage) {
	clientsMu.RLock()
	client, ok := clients[sessionID]
	clientsMu.RUnlock()

	if !ok || client == nil {
		return
	}

	msg.Timestamp = time.Now().Unix()

	client.mu.Lock()
	defer client.mu.Unlock()

	data, _ := json.Marshal(msg)
	if err := client.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("[CHAT] Error sending to %s: %v", sessionID, err)
	}
}

// SendAncientMessage sends a message from the Ancient One
func SendAncientMessage(sessionID, situation string) {
	msg := ChatMessage{
		Sender: "ancient_one",
		Text:   GetAncientQuip(situation),
		Type:   "speech",
	}
	SendToClient(sessionID, msg)
}

// ChatHandler handles WebSocket connections for chat
func ChatHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[CHAT] Upgrade error: %v", err)
		return
	}

	sessionID := getSessionID(w, r)
	log.Printf("[CHAT] Client connected: %s", sessionID)

	client := &ClientConnection{
		Conn:      conn,
		SessionID: sessionID,
	}

	clientsMu.Lock()
	clients[sessionID] = client
	clientsMu.Unlock()

	// Send greeting
	go func() {
		time.Sleep(500 * time.Millisecond)
		SendAncientMessage(sessionID, "greeting")
	}()

	// Read loop (for future player messages)
	defer func() {
		clientsMu.Lock()
		delete(clients, sessionID)
		clientsMu.Unlock()
		conn.Close()
		log.Printf("[CHAT] Client disconnected: %s", sessionID)
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		// For now, just log player messages
		log.Printf("[CHAT] From %s: %s", sessionID, string(message))
		// Future: Forward to LLM, parse commands, etc.
	}
}

// TriggerIdleMessage sends an impatient message after delay
// Call this when waiting for player input
func TriggerIdleMessage(sessionID string, delay time.Duration) *time.Timer {
	return time.AfterFunc(delay, func() {
		SendAncientMessage(sessionID, "idle")
	})
}
