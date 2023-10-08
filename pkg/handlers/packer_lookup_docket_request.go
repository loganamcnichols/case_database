package handlers

import (
	"log"
	"net/http"
)

func PacerLookupDocketRequest(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	docketNumber := r.FormValue("docket-number")
	caseID := r.FormValue("case-id")
	court := r.FormValue("court")

	log.Printf("Received docket number: %s, caseID: %s, court: %s", docketNumber, caseID, court)

}
