package handlers

import (
	"log"
	"net/http"
)

func PacerLookupDocketRequest(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	docketNumber := r.FormValue("docket-number")

	log.Printf("Received docket number: %s", docketNumber)
}
