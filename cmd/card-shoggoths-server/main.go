package main

import (
	"card-shoggoths/internal/server"
	"card-shoggoths/internal/store"
	"log"
	"net/http"
	"os"

	chi "github.com/go-chi/chi/v5"
)

func main() {
	// Init Store
	os.MkdirAll("./data", 0755)
	st, err := store.NewSQLiteStore("./data/game.db")
	if err != nil {
		log.Fatalf("Failed to init db: %v", err)
	}
	server.Init(st)

	r := chi.NewRouter()

	log.Println("Registering default static file handler")
	r.Handle("/*", http.FileServer(http.Dir("./static")))

	log.Println("Registering handlers...")
	r.HandleFunc("/api/deal", server.DealHandler)
	r.HandleFunc("/api/bet", server.ActionHandler) // Renamed from BetHandler
	r.HandleFunc("/api/discard", server.DiscardHandler)
	r.HandleFunc("/api/showdown", server.ShowdownHandler)
	r.HandleFunc("/debug/clear-session", server.ClearSessionHandler)

	log.Println("Serving on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
