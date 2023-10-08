package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/loganamcnichols/case_database/pkg/scraper"
)

var possibleCasesTemplate = template.Must(template.ParseFiles("web/templates/possible-cases.html"))

func PacerLookupOnSubmit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	court := r.FormValue("court")
	docket := r.FormValue("docket")

	log.Printf("Received court: %s, and docket number: %s", court, docket)
	url := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/possible_case_numbers.pl?%s", court, docket)

	client, err := scraper.LoginToPacer()
	if err != nil {
		log.Printf("Error logging in to PACER: %v", err)
		http.Error(w, "Error logging in to PACER", http.StatusInternalServerError)
		return
	}
	res, err := scraper.PossbleCasesSearch(client, url)
	if err != nil {
		log.Printf("Error searching for possible cases: %v", err)
		http.Error(w, "Error searching for possible cases", http.StatusInternalServerError)
		return
	}
	possibleCasesTemplate.Execute(w, res)

}
