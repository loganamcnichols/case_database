package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/loganamcnichols/case_database/pkg/db"
	"github.com/loganamcnichols/case_database/pkg/scraper"
)

type TemplateData struct {
	Court string
	Cases scraper.PossibleCases
}

func PacerLookupOnSubmit(w http.ResponseWriter, r *http.Request) {
	var possibleCasesTemplate = template.Must(template.ParseFiles("web/templates/possible-cases.html"))
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

	templateData := TemplateData{
		Court: court,
		Cases: res,
	}
	possibleCasesTemplate.Execute(w, templateData)
	cnx, err := db.Connect()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		return
	}
	defer cnx.Close()
	for _, c := range res.Cases {
		log.Printf("Case ID: %s, Title: %s", c.ID, c.Title)
		caseID, err := strconv.Atoi(c.ID)
		if err != nil {
			log.Printf("Error converting case ID to int: %v", err)
			return
		}
		err = db.InsertCases(cnx, court, caseID, c.Title, c.Number)
		if err != nil {
			log.Printf("Error inserting case into database: %v", err)
			return
		}
	}
}
