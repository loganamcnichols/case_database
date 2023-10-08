package scraper

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

const LoginURL = "https://pacer.login.uscourts.gov/services/cso-auth"

func LoginToPacer() (*http.Client, error) {

	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file")
	}
	// Fetch credentials.
	username := os.Getenv("PACER_USERNAME")
	password := os.Getenv("PACER_PASSWORD")
	nextGenCSO := os.Getenv("NextGenCSO")

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: time.Second * 10, // Making the timeout explicit as 10 seconds
		Jar:     jar,
	}

	// Set the cookie.
	cookie := &http.Cookie{
		Name:   "NextGenCSO",
		Value:  nextGenCSO,
		Domain: "uscourts.gov",
		Path:   "/",
	}
	u, _ := url.Parse(LoginURL)
	jar.SetCookies(u, []*http.Cookie{cookie})

	// Try accessing pacer resource with cookie
	data, err := PossbleCasesSearch(client, "https://ecf.azd.uscourts.gov/cgi-bin/possible_case_numbers.pl?22-02189")
	if err == nil && len(data.Cases) > 0 {
		// CSO token still valid.
		return client, nil
	}

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

	if err != nil {
		return nil, err
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
	// Set the cookie, and write it to the .env file
	os.Setenv("NextGenCSO", pacerResp.NextGenCSO)
	cookie = &http.Cookie{
		Name:   "NextGenCSO",
		Value:  pacerResp.NextGenCSO,
		Domain: "uscourts.gov",
		Path:   "/",
	}
	// Set the OS environment variable
	appendToEnvFile("NextGenCSO", pacerResp.NextGenCSO)
	u, _ = url.Parse(LoginURL)
	jar.SetCookies(u, []*http.Cookie{cookie})
	return client, nil
}

type PossibleCases struct {
	Number string         `xml:"number,attr"`
	Cases  []PossibleCase `xml:"case"`
}

type PossibleCase struct {
	Number   string `xml:"number,attr"`
	ID       string `xml:"id,attr"`
	Title    string `xml:"title,attr"`
	Sortable string `xml:"sortable,attr"`
}

func PossbleCasesSearch(client *http.Client, url string) (PossibleCases, error) {
	var respStruct PossibleCases
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

func GetDownloadLink(client *http.Client, url string, referer string, docNo int, caseNum int) (string, error) {
	var downloadLink string
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	field1, err := writer.CreateFormField(fmt.Sprintf("CaseNum_%d", caseNum))
	if err != nil {
		return downloadLink, err
	}
	field1.Write([]byte("on"))

	field2, err := writer.CreateFormField("document_number")
	if err != nil {
		return downloadLink, err
	}
	field2.Write([]byte(strconv.Itoa(docNo)))
	writer.Close()

	req, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return downloadLink, err
	}
	if err != nil {
		return downloadLink, err
	}
	req.Header.Set("User-Agent", "loganamcnichols")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Referer", referer)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return downloadLink, err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return downloadLink, err
	}
	fmt.Println(string(bodyBytes))

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return downloadLink, fmt.Errorf("recieved non found code")
	}
	u := resp.Request.URL
	u.RawQuery = ""
	downloadLink = u.String()

	return downloadLink, nil

}

func appendToEnvFile(key, value string) error {
	// Open the .env file with flags to append data and create the file if it doesn't exist
	file, err := os.OpenFile(".env", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the new key-value pair to the file
	_, err = file.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	return err
}

func GetDocketSummaryLink(doc goquery.Document) string {
	var docketSummaryLink string
	doc.Find("table").First().Find("a").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "Docket Report") {
			docketSummaryLink, _ = s.Attr("href")
		}
	})
	return docketSummaryLink
}

func GetCaseMainPage(client *http.Client, url string, case_id int, case_number string) (*goquery.Document, error) {
	var document *goquery.Document
	buffer := bytes.Buffer{}
	writer := multipart.NewWriter(&buffer)

	_, err := writer.CreateFormField("UserType")
	if err != nil {
		return document, err
	}

	field2, err := writer.CreateFormField("all_case_ids")
	if err != nil {
		return document, err
	}
	field2.Write([]byte(strconv.Itoa(case_id)))
	field3, err := writer.CreateFormField("case_num")
	if err != nil {
		return document, err
	}
	field3.Write([]byte(case_number))
	_, err = writer.CreateFormField("Qry_filed_from")
	if err != nil {
		return document, err
	}
	_, err = writer.CreateFormField("Qry_filed_to")
	if err != nil {
		return document, err
	}
	_, err = writer.CreateFormField("lastentry_from")
	if err != nil {
		return document, err
	}
	_, err = writer.CreateFormField("lastentry_to")
	if err != nil {
		return document, err
	}
	_, err = writer.CreateFormField("last_name")
	if err != nil {
		return document, err
	}
	_, err = writer.CreateFormField("first_name")
	if err != nil {
		return document, err
	}
	_, err = writer.CreateFormField("middle_name")
	if err != nil {
		return document, err
	}
	_, err = writer.CreateFormField("person_type")
	if err != nil {
		return document, err
	}
	writer.Close()

	req, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return document, err
	}

	req.Header.Set("User-Agent", "loganamcnichols")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Origin", "https://ecf.almd.uscourts.gov")
	req.Header.Set("Referer", "https://ecf.almd.uscourts.gov/cgi-bin/iquery.pl")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return document, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return document, fmt.Errorf("recieved non found code")
	}
	document, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return document, err
	}
	return document, nil
}
