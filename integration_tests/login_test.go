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

	_, err := scraper.LoginToPacer(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Check if login was successful
	var xpathSelector = `//*[contains(text(), "Logan McNichols")]`

	var outerHTML string
	err = chromedp.Run(ctx,
		chromedp.OuterHTML(xpathSelector, &outerHTML, chromedp.BySearch))

	if err != nil {
		t.Fatal(err)
	}
	t.Log(outerHTML)
}
