package integrationtests

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/loganamcnichols/case_database/pkg/scraper"
)

var client *http.Client

func TestMain(m *testing.M) {
	// setup code
	var err error
	client, err = scraper.LoginToPacer()
	if err != nil {
		fmt.Println("Error logging in to PACER")
		os.Exit(1)
	}
	m.Run()
}

func TestLoginToPacer(t *testing.T) {
	client, err := scraper.LoginToPacer()
	if err != nil {
		t.Fatalf("LoginToPacer() returned error: %v", err)
	}
	u, err := url.Parse(scraper.LoginURL)
	if err != nil {
		t.Fatalf("url.Parse() returned error: %v", err)
	}
	cookies := client.Jar.Cookies(u)
	cookieName := "NextGenCSO"
	for _, cookie := range cookies {
		if cookie.Name == cookieName {
			return
		}
	}
	t.Fatalf("LoginToPacer() did not return a %s cookie", cookieName)

}

func TestSearchByDocketNumber(t *testing.T) {
	client, _ := scraper.LoginToPacer()
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
	count, err := scraper.DocketCountFromCaseId(baseURL, client, 1313500)
	if err != nil {
		t.Fatalf("DocketCountFromCaseId() returned error: %v", err)
	}
	if count != 25 {
		t.Fatalf("DocketCountFromCaseId() returned incorrect docket count: %d", count)
	}
}

func TestGetDownloadLink(t *testing.T) {
	requestUrl := "https://ecf.almd.uscourts.gov/cgi-bin/qryDocument.pl?644448178352274-L_1_0-1"
	referer := "https://ecf.almd.uscourts.gov/cgi-bin/qryDocument.pl?56135"
	expectedResponseURL := "https://ecf.almd.uscourts.gov/doc1/01712410676"
	responseURL, err := scraper.GetDownloadLink(client, requestUrl, referer, 1, 72385)
	if err != nil {
		t.Fatalf("GetDownloadLink() returned error: %v", err)
	}
	if responseURL != expectedResponseURL {
		t.Fatalf("GetDownloadLink() returned incorrect response URL: %s", responseURL)
	}
}

// func TestGetDocketSummaryLink(t *testing.T) {
// 	requestURL := "https://ecf.almd.uscourts.gov/cgi-bin/iquery.pl?13573439176722-L_1_0-1"
// 	expectedResponseURL := "https://ecf.almd.uscourts.gov/cgi-bin/DktRpt.pl?56135"
// 	responseURL, err := scraper.GetDocketSummaryLink(client, requestURL)
// 	if err != nil {
// 		t.Fatalf("GetDocketSummaryLink() returned error: %v", err)
// 	}
// 	if responseURL != expectedResponseURL {
// 		t.Fatalf("GetDocketSummaryLink() returned incorrect response URL: %s", responseURL)
// 	}
// }

func TestGetCaseMainPage(t *testing.T) {
	requestURL := "https://ecf.almd.uscourts.gov/cgi-bin/iquery.pl?154632979339918-L_1_0-1"
	document, err := scraper.GetCaseMainPage(client, requestURL, 56135, "2:14-cr-646")
	if err != nil {
		t.Fatalf("GetCaseMainPage() returned error: %v", err)
	}
	docText := document.Find("body").Text()
	if !strings.Contains(docText, "USA v. Manniken") {
		t.Fatalf("GetCaseMainPage() returned incorrect document: %s", docText)
	}
}

func TestGetDocketSummaryLink(t *testing.T) {
	requestURL := "https://ecf.almd.uscourts.gov/cgi-bin/iquery.pl?154632979339918-L_1_0-1"
	document, err := scraper.GetCaseMainPage(client, requestURL, 56135, "2:14-cr-646")
	if err != nil {
		t.Fatalf("GetCaseMainPage() returned error: %v", err)
	}
	queryPage := scraper.GetDocketSummaryLink(*document)
	if queryPage != "/cgi-bin/DktRpt.pl?56135" {
		t.Fatalf("GetDocketSummaryLink() returned incorrect URL: %s", queryPage)
	}
}
