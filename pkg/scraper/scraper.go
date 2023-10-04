package scraper

import (
	"context"
	"os"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func LoginToPacer(ctx context.Context) (*network.Response, error) {
	// Fetch credentials.
	username := os.Getenv("PACER_USERNAME")
	password := os.Getenv("PACER_PASSWORD")

	// Define login tasks.
	tasks := chromedp.Tasks{
		chromedp.Navigate(`https://pacer.login.uscourts.gov/csologin/login.jsf`),
		chromedp.WaitVisible(`#loginForm\:fbtnLogin`, chromedp.ByID),
		chromedp.SendKeys(`#loginForm\:loginName`, username, chromedp.ByID),
		chromedp.SendKeys(`#loginForm\:password`, password, chromedp.ByID),
		chromedp.Click(`#loginForm\:fbtnLogin`, chromedp.ByID),
	}
	return chromedp.RunResponse(ctx, tasks)
}
