// handler.go

package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/loganamcnichols/case_database/pkg/db"
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
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	caseID := r.URL.Query().Get("caseID")
	court := r.URL.Query().Get("court")
	caseNumber := r.URL.Query().Get("caseNumber")

	fmt.Printf("Received caseID: %s, caseNumber: %s\n, courtID: %s\n", caseID, caseNumber, court)

	nextGenCSO, _ := r.Cookie("NextGenCSO")
	client, err := scraper.LoginToPacer("", "", nextGenCSO.Value)
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

	con, err := db.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := con.Query("SELECT doc_number, description FROM documents WHERE case_id = $1", id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var docNumber int
		var description string
		err = rows.Scan(&docNumber, &description)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Fprintf(w, "DocNumber: %d, Description: %s\n", docNumber, description)
	}

}
