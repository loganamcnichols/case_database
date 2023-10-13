package integrationtests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/loganamcnichols/case_database/pkg/db"
	"github.com/loganamcnichols/case_database/pkg/handlers"
	"github.com/loganamcnichols/case_database/pkg/scraper"
)

var client *http.Client

func TestMain(m *testing.M) {
	// setup code
	var err error
	username := os.Getenv("PACER_USERNAME")
	password := os.Getenv("PACER_PASSWORD")
	token := os.Getenv("NextGenCSO")
	client, err = scraper.LoginToPacer(username, password, token)
	u, _ := url.Parse(scraper.LoginURL)
	cookies := client.Jar.Cookies(u)
	for _, cookie := range cookies {
		if cookie.Name == "NextGenCSO" {
			os.Setenv("NextGenCSO", cookie.Value)
		}
	}

	if err != nil {
		fmt.Println("Error logging in to PACER")
		os.Exit(1)
	}
	m.Run()
}

func TestSearchByDocketNumber(t *testing.T) {
	data, err := scraper.PossbleCasesSearch(client, "https://ecf.azd.uscourts.gov/cgi-bin/possible_case_numbers.pl?22-02189")
	if err != nil {
		t.Fatalf("SearchByDocketNumber() returned error: %v", err)
	}
	if data.Number != "22-02189" {
		t.Fatalf("SearchByDocketNumber() returned incorrect docket number: %s", data.Number)
	}
	if len(data.Cases) != 4 {
		t.Fatalf("SearchByDocketNumber() returned incorrect number of cases: %d", len(data.Cases))
	}
	if data.Cases[2].ID != "1313500" {
		t.Fatalf("SearchByDocketNumber() returned incorrect case ID: %s", data.Cases[2].ID)
	}
}

func TestDocketCountFromCaseId(t *testing.T) {
	baseURL := "https://ecf.azd.uscourts.gov/cgi-bin/mobile_query.pl"
	refererURL := "https://ecf.azd.uscourts.gov/cgi-bin/iquery.pl"
	count, err := scraper.DocketCountFromCaseId(baseURL, refererURL, client, "1313500")
	if err != nil {
		t.Fatalf("DocketCountFromCaseId() returned error: %v", err)
	}
	if count != 25 {
		t.Fatalf("DocketCountFromCaseId() returned incorrect docket count: %d", count)
	}
}

func TestGetDocketSummaryLink(t *testing.T) {
	requestURL := "https://ecf.almd.uscourts.gov/cgi-bin/iquery.pl?154632979339918-L_1_0-1"
	document, err := scraper.GetCaseMainPage(client, requestURL, "56135", "2:14-cr-646")
	if err != nil {
		t.Fatalf("GetCaseMainPage() returned error: %v", err)
	}
	queryPage := scraper.GetDocketSummaryLink(*document)
	if queryPage != "/cgi-bin/DktRpt.pl?56135" {
		t.Fatalf("GetDocketSummaryLink() returned incorrect URL: %s", queryPage)
	}
}

func TestGetCaseURL(t *testing.T) {
	requestURL := "https://ecf.almd.uscourts.gov/cgi-bin/iquery.pl"
	respURL, err := scraper.GetFormURL(client, requestURL)
	if err != nil {
		t.Fatalf("GetCaseURL() returned error: %v", err)
	}
	document, err := scraper.GetCaseMainPage(client, respURL, "56135", "2:14-cr-646")
	if err != nil {
		t.Fatalf("GetCaseMainPage() returned error: %v", err)
	}
	docText := document.Find("body").Text()
	if !strings.Contains(docText, "USA v. Manniken") {
		t.Fatalf("GetCaseMainPage() returned incorrect document: %s", docText)
	}
}

func TestGetDocumentURL(t *testing.T) {
	requestURL := "https://ecf.almd.uscourts.gov/cgi-bin/qryDocument.pl?56135"
	expectedResponseURL := "https://ecf.almd.uscourts.gov/doc1/01712410676"
	respURL, err := scraper.GetFormURL(client, requestURL)
	if err != nil {
		t.Fatalf("GetDocumentURL() returned error: %v", err)
	}
	downLoadLink, err := scraper.GetDownloadLink(client, respURL, requestURL, 1, 72385)
	if err != nil {
		t.Fatalf("GetDocumentURL() returned error: %v", err)
	}
	if downLoadLink != expectedResponseURL {
		t.Fatalf("GetDocumentURL() returned incorrect URL: %s", downLoadLink)
	}
}

func TestPacerLookup(t *testing.T) {
	os.Chdir("../")
	// Create form data
	formData := url.Values{}
	formData.Add("court", "azd") // Sample value
	formData.Add("docket", "22-02189")

	token := os.Getenv("NextGenCSO")

	cookie := &http.Cookie{
		Name:   "NextGenCSO",
		Value:  token,
		Domain: "uscourts.gov",
		Path:   "/",
	}

	// Create a request to pass to the handler
	req, err := http.NewRequest("POST", "/pacer-lookup-submit", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatal(err)
	}

	// Set the header for form data
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.AddCookie(cookie)
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

}
