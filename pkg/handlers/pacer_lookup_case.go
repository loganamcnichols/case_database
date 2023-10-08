// handler.go

package handlers

import (
	"fmt"
	"net/http"
)

func PacerLookupCase(w http.ResponseWriter, r *http.Request) {
	// For hx-get with hx-vals, values are sent as query parameters
	caseID := r.URL.Query().Get("caseID")
	court := r.URL.Query().Get("court")
	caseNumber := r.URL.Query().Get("caseNumber")

	fmt.Printf("Received caseID: %s, caseNumber: %s\n, courtID: %s\n", caseID, caseNumber, court)

	//		client, err := scraper.LoginToPacer()
	//		if err != nil {
	//			http.Error(w, "Error logging in to PACER", http.StatusInternalServerError)
	//			return
	//		}
	//		baseURL := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/mobile_query.pl",
	//	}
}
