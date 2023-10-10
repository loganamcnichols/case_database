package integrationtests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/loganamcnichols/case_database/pkg/db"
	"github.com/loganamcnichols/case_database/pkg/handlers"
)

func TestPacerLookup(t *testing.T) {
	os.Chdir("../")
	// Create form data
	formData := url.Values{}
	formData.Add("court", "azd") // Sample value
	formData.Add("docket", "22-02189")

	// Create a request to pass to the handler
	req, err := http.NewRequest("POST", "/pacer-lookup-submit", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	// Set the header for form data
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Other test steps...
	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Create a handler function
	handler := http.HandlerFunc(handlers.PacerLookupOnSubmit)

	// Call the handler function
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v, expected %v", status, http.StatusOK)
	}

	cnx, err := db.Connect()
	if err != nil {
		t.Fatalf("Error connecting to database: %v", err)
	}
	defer cnx.Close()

	if err != nil {
		t.Errorf("Error beginning transaction: %v", err)
	}

	if err != nil {
		t.Fatalf("Error connecting to database: %v", err)
	}
	defer cnx.Close()
	cases, err := db.QueryCases(cnx, "azd", 1312364)
	if err != nil {
		t.Fatalf("Error querying casecnx %v", err)
	}
	if len(cases) == 0 {
		t.Fatalf("QueryCases() returned no cases")
	}
	cnx.Exec("DELETE FROM cases WHERE id != 1")

	// Check the response body, headers, etc. as required for your test
	// Example:
	// expected := `Your expected response`
	// if rr.Body.String() != expected {
	//     t.Errorf("Handler returned unexpected body: got %v, expected %v", rr.Body.String(), expected)
	// }

}
