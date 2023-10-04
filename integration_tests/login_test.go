package integrationtests

import (
	"context"
	"testing"

	"github.com/loganamcnichols/case_database/pkg/scraper"

	"github.com/chromedp/chromedp"
)

func TestLoginToPacer(t *testing.T) {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Check if we are logged in, should be false.
	var loggedIn bool
	loggedIn, err := scraper.LoggedIn(ctx)
	if err != nil {
		t.Fatal(err)
	} else if loggedIn {
		t.Fatal("false login signal")
	}

	// Log in.
	_, err = scraper.LoginToPacer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Check if we are logged in. Should be true.
	loggedIn, err = scraper.LoggedIn(ctx)
	if err != nil {
		t.Fatal(err)
	} else if !loggedIn {
		t.Fatal("should be logged in")
	}
}
