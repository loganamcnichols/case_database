package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/loganamcnichols/case_database/pkg/handlers"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/pacer-lookup", handlers.PacerLookupHandler).Methods("GET") // Add this line

	http.Handle("/", r)
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
