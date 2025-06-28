package main

import (
	"log"
	"net/http"

	"card-shoggoths/internal/server"
)

func main() {
	log.Println("Registering default static file handler")
	http.Handle("/", http.FileServer(http.Dir("./static")))
	log.Println("Registering handler /api/deal")
	http.HandleFunc("/api/deal", server.DealHandler)
	log.Println("Registering handler /api/discard");
	http.HandleFunc("/api/discard", server.DiscardHandler)
	log.Println("Registering handler /api/showdown")
	http.HandleFunc("/api/showdown", server.ShowdownHandler)

	http.HandleFunc("/debug/clear-session", server.ClearSessionHandler)
	log.Println("Serving on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}