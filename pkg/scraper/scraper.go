package scraper

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
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
	defer resp.Body.Close()

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return client, fmt.Errorf("received non-2xx response status: %d %s", resp.StatusCode, resp.Status)
	}

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
		Name:   "NextGenCSO",
		Value:  pacerResp.NextGenCSO,
		Domain: "uscourts.gov",
		Path:   "/",
	}
	u, _ := url.Parse(LoginURL)
	jar.SetCookies(u, []*http.Cookie{cookie})
	return client, nil
}

type CaseNumberResponse struct {
	Number string            `xml:"number,attr"`
	Cases  []CaseNumberEntry `xml:"case"`
}

type CaseNumberEntry struct {
	Number   string `xml:"number,attr"`
	ID       string `xml:"id,attr"`
	Title    string `xml:"title,attr"`
	Sortable string `xml:"sortable,attr"`
}

func SearchByDocketNumber(client *http.Client, url string) (CaseNumberResponse, error) {
	var respStruct CaseNumberResponse
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return respStruct, err
	}
	req.Header.Set("User-Agent", "loganamcnichols")
	req.Header.Set("Accept", "application/xml")

	resp, err := client.Do(req)
	if err != nil {
		return respStruct, err
	}
	defer resp.Body.Close()

	// Check for non-2xx status codes
	if resp.StatusCode != http.StatusOK {
		return respStruct, fmt.Errorf("received non-2xx response status: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	fmt.Println(string(body))
	if err != nil {
		return respStruct, fmt.Errorf("failed to read response body: %v", err)
	}
	err = xml.Unmarshal(body, &respStruct)
	if err != nil {
		return respStruct, fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	return respStruct, nil
}

func DocketCountFromCaseId(baseURL string, client *http.Client, id int) (int, error) {
	var docketCount int
	u, err := url.Parse(baseURL)
	if err != nil {
		return docketCount, err
	}
	q := u.Query()
	q.Set("search", "caseInfo")
	q.Set("caseid", strconv.Itoa(id))
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return docketCount, err
	}

	req.Header.Set("User-Agent", "loganamcnichols")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Referer", "https://ecf.azd.uscourts.gov/cgi-bin/iquery.pl")

	resp, err := client.Do(req)
	if err != nil {
		return docketCount, err
	}
	defer resp.Body.Close()

	// body, _ := io.ReadAll(resp.Body)

	// fmt.Println(string(body))
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return docketCount, err
	}
	fmt.Println(document.Text())
	spanElement := document.Find("a#entriesLink").First()
	re := regexp.MustCompile("[0-9]+")
	digitString := re.FindString(spanElement.Text())

	// Convert string of digits to integer
	docketCount, err = strconv.Atoi(digitString)
	if err != nil {
		return docketCount, err
	}
	return docketCount, nil

}
