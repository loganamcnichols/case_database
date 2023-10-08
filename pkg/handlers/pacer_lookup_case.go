// handler.go

package handlers

import (
	"fmt"
	"net/http"
)

func PacerLookupCase(w http.ResponseWriter, r *http.Request) {
	// For hx-get with hx-vals, values are sent as query parameters
	caseID := r.URL.Query().Get("caseID")
	caseNumber := r.URL.Query().Get("CaseNumber")

	fmt.Printf("Received caseID: %s, caseNumber: %s\n", caseID, caseNumber)
}
