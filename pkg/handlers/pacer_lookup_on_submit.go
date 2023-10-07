package handlers

import (
	"log"
	"net/http"
)

func PacerLookupOnSubmit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	court := r.FormValue("court")
	docket := r.FormValue("docket")

	log.Printf("Received court: %s, and docket number: %s", court, docket)

	w.Write([]byte("Data received"))
}
