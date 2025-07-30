package main

import (
	"card-shoggoths/internal/server"
	"log"
	"net/http"

	chi "github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	log.Println("Registering default static file handler")
	r.Handle("/*", http.FileServer(http.Dir("./static")))

	log.Println("Registering handler /api/deal")
	r.HandleFunc("/api/deal", server.DealHandler)

	log.Println("Registering handler /api/bet")
	r.HandleFunc("/api/bet", server.BetHandler)

	log.Println("Registering handler /api/discard")
	r.HandleFunc("/api/discard", server.DiscardHandler)

	log.Println("Registering handler /api/showdown")
	r.HandleFunc("/api/showdown", server.ShowdownHandler)

	log.Println("Registering debug handler /debug/clear-session")
	r.HandleFunc("/debug/clear-session", server.ClearSessionHandler)

	log.Println("Serving on :8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
