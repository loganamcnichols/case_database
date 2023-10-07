package integrationtests

import (
	"net/url"
	"testing"

	"github.com/loganamcnichols/case_database/pkg/scraper"
)

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
	data, err := scraper.SearchByDocketNumber(client, "https://ecf.azd.uscourts.gov/cgi-bin/possible_case_numbers.pl?22-02189")
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
