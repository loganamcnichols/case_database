package scraper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

const LoginURL = "https://pacer.login.uscourts.gov/services/cso-auth"

func LoginToPacer() (*http.Client, error) {
	// Fetch credentials.
	username := os.Getenv("PACER_USERNAME")
	password := os.Getenv("PACER_PASSWORD")

	// Check for empty credentials
	if username == "" || password == "" {
		return nil, errors.New("PACER_USERNAME or PACER_PASSWORD environment variables are not set")
	}

	// Create request.
	jsonBody := []byte(fmt.Sprintf(`{"loginId":"%s","password":"%s","redactFlag":"1"}`, username, password))
	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest("POST", LoginURL, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", username)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: time.Second * 10, // Making the timeout explicit as 10 seconds
		Jar:     jar,
	}
	resp, err := client.Do(req)
	if err != nil {
		return client, err
	}

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return client, fmt.Errorf("received non-2xx response status: %d %s", resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()

	pacerResp := struct {
		ErrorDescription string `json:"errorDescription"`
		NextGenCSO       string `json:"nextGenCSO"`
	}{}
	// Check for errors or an empty NextGenCSO cookie.
	if err := json.NewDecoder(resp.Body).Decode(&pacerResp); err != nil {
		return client, fmt.Errorf("failed to decode response body: %v", err)
	} else if pacerResp.ErrorDescription != "" {
		return client, fmt.Errorf("error from PACER authentication: %s", pacerResp.ErrorDescription)
	} else if pacerResp.NextGenCSO == "" {
		return client, fmt.Errorf("no NextGenCSO cookie found in response")
	}
	// Set the cookie.
	cookie := &http.Cookie{
		Name:   "nextGenCSO",
		Value:  pacerResp.NextGenCSO,
		Domain: "uscourts.gov",
		Path:   "/",
	}
	u, _ := url.Parse(LoginURL)
	jar.SetCookies(u, []*http.Cookie{cookie})
	return client, nil
}
