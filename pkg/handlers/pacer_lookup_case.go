// handler.go

package handlers

import (
	"fmt"
	"net/http"

	"html/template"

	"github.com/loganamcnichols/case_database/pkg/scraper"
)

type DocumentTemplateData struct {
	Count  int
	CaseID string
	Court  string
}

func PacerLookupCase(w http.ResponseWriter, r *http.Request) {
	var docketNumber = template.Must(template.ParseFiles("web/templates/docket-number.html"))
	// For hx-get with hx-vals, values are sent as query parameters
	caseID := r.URL.Query().Get("caseID")
	court := r.URL.Query().Get("court")
	caseNumber := r.URL.Query().Get("caseNumber")

	fmt.Printf("Received caseID: %s, caseNumber: %s\n, courtID: %s\n", caseID, caseNumber, court)

	client, err := scraper.LoginToPacer()
	if err != nil {
		http.Error(w, "Error logging in to PACER", http.StatusInternalServerError)
		return
	}
	baseURL := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/iquery.pl", court)
	moblileURL := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/mobile_query.pl", court)
	caseURL, err := scraper.GetFormURL(client, baseURL)
	if err != nil {
		http.Error(w, "Error getting case URL", http.StatusInternalServerError)
		return
	}
	casePage, err := scraper.GetCaseMainPage(client, caseURL, caseID, caseNumber)
	if err != nil {
		http.Error(w, "Error getting case page", http.StatusInternalServerError)
		return
	}

	metadata := casePage.Find("center")

	metadataHTML, err := metadata.Html()
	if err != nil {
		http.Error(w, "Error getting case metadata", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, metadataHTML)

	count, err := scraper.DocketCountFromCaseId(moblileURL, caseURL, client, caseID)
	if err != nil {
		http.Error(w, "Error getting docket count", http.StatusInternalServerError)
		return
	}
	templateData := DocumentTemplateData{
		Count:  count,
		CaseID: caseID,
		Court:  court,
	}
	fmt.Fprintf(w, "Docket count: %d\n", count)
	docketNumber.Execute(w, templateData)
}
