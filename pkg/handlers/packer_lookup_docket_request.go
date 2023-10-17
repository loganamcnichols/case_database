package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/loganamcnichols/case_database/pkg/db"
	"github.com/loganamcnichols/case_database/pkg/scraper"
)

func PacerLookupDocketRequest(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	docketNumber := r.FormValue("docket-number")
	caseID := r.FormValue("case-id")
	court := r.FormValue("court")

	nextGenCSO, _ := r.Cookie("NextGenCSO")
	userID := CheckSession(r)

	client, err := scraper.LoginToPacer("", "", nextGenCSO.Value)
	if err != nil {
		log.Printf("Error logging in to PACER: %v", err)
		http.Error(w, "Error logging in to PACER", http.StatusInternalServerError)
		return
	}

	requestURL := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/qryDocument.pl?%s", court, caseID)

	respURL, err := scraper.GetFormURL(client, requestURL)
	if err != nil {
		log.Printf("Error getting form URL: %v", err)
		http.Error(w, "Error getting form URL", http.StatusInternalServerError)
		return
	}
	downloadLink, deSeqNum, err := scraper.GetDownloadLinks(client, respURL, requestURL, docketNumber, caseID)
	if err != nil {
		log.Printf("Error getting download link: %v", err)
		http.Error(w, "Error getting download link", http.StatusInternalServerError)
		return
	}
	log.Printf("Received docket number: %s, caseID: %s, court: %s", docketNumber, caseID, court)
	resDoc, err := scraper.PurchaseDocument(client, downloadLink[0], caseID, deSeqNum)
	fmt.Println(resDoc.Find("body").Text())
	if err != nil {
		log.Printf("Error purchasing document: %v", err)
		http.Error(w, "Error purchasing document", http.StatusInternalServerError)
		return
	}
	_, err = scraper.PerformDownload(client, resDoc, downloadLink[0], caseID, docketNumber)
	if err != nil {
		log.Printf("Error performing download: %v", err)
		http.Error(w, "Error performing download", http.StatusInternalServerError)
		return
	}

	cnx, err := db.Connect()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		return
	}
	defer cnx.Close()
	pages, err := scraper.GetPageCount(client, downloadLink, respURL)
	if err != nil {
		log.Printf("Error getting page count: %v", err)
		http.Error(w, "Error getting page count", http.StatusInternalServerError)
		return
	}

	cnx.Exec(`INSERT INTO documents (description, file, doc_number, case_id, pages, user_id)
				VALUES ('description', $1, $2, $3, $4, $5)`, docketNumber, docketNumber, caseID, pages, userID)

}

func PacerLookupSummaryRequest(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	caseID := r.FormValue("case-id")
	court := r.FormValue("court")

	nextGenCSO, _ := r.Cookie("NextGenCSO")

	client, err := scraper.LoginToPacer("", "", nextGenCSO.Value)
	if err != nil {
		log.Printf("Error logging in to PACER: %v", err)
		http.Error(w, "Error logging in to PACER", http.StatusInternalServerError)
		return
	}

	requestURL := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/DktRpt.pl?%s", court, caseID)
	respURL, err := scraper.GetFormURL(client, requestURL)
	if err != nil {
		log.Printf("Error getting form URL: %v", err)
		http.Error(w, "Error getting form URL", http.StatusInternalServerError)
		return
	}
	document, err := scraper.GetDocumentSummary(client, respURL, caseID)
	if err != nil {
		log.Printf("Error getting document summary: %v", err)
		http.Error(w, "Error getting document summary", http.StatusInternalServerError)
		return
	}

	data, err := document.Find("#cmecfMainContent").First().Html()
	if err != nil {
		log.Printf("Error getting document summary: %v", err)
		http.Error(w, "Error getting document summary", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, data)
}
