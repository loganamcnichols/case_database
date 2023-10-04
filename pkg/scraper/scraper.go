package scraper

import (
	"context"
	"os"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func LoginToPacer(ctx context.Context) (*network.Response, error) {
	// Fetch credentials.
	username := os.Getenv("PACER_USERNAME")
	password := os.Getenv("PACER_PASSWORD")

	// Define login tasks.
	tasks := chromedp.Tasks{
		chromedp.WaitVisible(`#loginForm\:fbtnLogin`, chromedp.ByID),
		chromedp.SendKeys(`#loginForm\:loginName`, username, chromedp.ByID),
		chromedp.SendKeys(`#loginForm\:password`, password, chromedp.ByID),
		chromedp.Click(`#loginForm\:fbtnLogin`, chromedp.ByID),
	}
	return chromedp.RunResponse(ctx, tasks)
}

func LoggedIn(ctx context.Context) (bool, error) {
	loggedIn := false

	// Check if we are on the login page.
	var url string
	err := chromedp.Run(ctx, chromedp.Location(&url))
	if err != nil {
		return false, err
	}
	if url != "https://pacer.login.uscourts.gov/csologin/login.jsf" {
		chromedp.RunResponse(ctx, chromedp.Navigate(`https://pacer.login.uscourts.gov/csologin/login.jsf`))
	}

	// Check if we are logged in.
	var xpathSelector = `//*[contains(text(), "Logan McNichols")]`
	var nodes []*cdp.Node
	err = chromedp.Run(ctx,
		chromedp.Nodes(xpathSelector, &nodes, chromedp.BySearch, chromedp.AtLeast(0)))
	if err != nil {
		return loggedIn, err
	}
	if len(nodes) > 0 {
		loggedIn = true
	} else {
		loggedIn = false
	}
	return loggedIn, nil
}
