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
	cookieName := "nextGenCSO"
	for _, cookie := range cookies {
		if cookie.Name == cookieName {
			return
		}
	}
	t.Fatalf("LoginToPacer() did not return a %s cookie", cookieName)

}
