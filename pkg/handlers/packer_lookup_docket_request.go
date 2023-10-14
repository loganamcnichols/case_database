package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/loganamcnichols/case_database/pkg/scraper"
)

func PacerLookupDocketRequest(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	docketNumber := r.FormValue("docket-number")
	caseID := r.FormValue("case-id")
	court := r.FormValue("court")

	client, err := scraper.LoginToPacer("", "", nextGenCSO.Value)
	if err != nil {
		log.Printf("Error logging in to PACER: %v", err)
		http.Error(w, "Error logging in to PACER", http.StatusInternalServerError)
		return
	}

	requestURL := fmt.Sprintf("https://ecf.almd.uscourts.gov/cgi-bin/qryDocument.pl?%s", caseID)

	respURL, err := scraper.GetFormURL(client, requestURL)
	downloadLink, err := scraper.GetDownloadLink(client, respURL, requestURL, docketNumber, caseID)
	log.Printf("Received docket number: %s, caseID: %s, court: %s", docketNumber, caseID, court)
	resDoc, err := scraper.PurchaseDocument(client, downloadLink, "1348139", "9")
	if err != nil {
		log.Printf("Error purchasing document: %v", err)
		http.Error(w, "Error purchasing document", http.StatusInternalServerError)
		return
	}
	err = scraper.PerformDownload(client, resDoc, downloadLink, "1348139", "9")
}
