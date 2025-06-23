package main

import (
	"log"
	"net/http"

	"card-shoggoths/internal/server"
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/api/deal", server.DealHandler)
	http.HandleFunc("/api/discard", server.DiscardHandler)
	http.HandleFunc("/api/showdown", server.ShowdownHandler)

	log.Println("Serving on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
