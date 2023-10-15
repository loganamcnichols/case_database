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
	"time"

	"github.com/PuerkitoBio/goquery"
)

const LoginURL = "https://pacer.login.uscourts.gov/services/cso-auth"

func LoginToPacer(username string, password string, token string) (*http.Client, error) {
	// Prepare the cookie jar.
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: time.Second * 20, // Making the timeout explicit as 10 seconds
		Jar:     jar,
	}
	// Try first with provided token.
	if token != "" {
		// Set the cookie.
		cookie := &http.Cookie{
			Name:   "NextGenCSO",
			Value:  token,
			Domain: "uscourts.gov",
			Path:   "/",
		}
		u, _ := url.Parse(LoginURL)
		jar.SetCookies(u, []*http.Cookie{cookie})
		data, err := PossbleCasesSearch(client, "https://ecf.azd.uscourts.gov/cgi-bin/possible_case_numbers.pl?22-02189")
		if err == nil && len(data.Cases) > 0 {
			// CSO token still valid.
			return client, nil
		}
	}
	// Bail early if we don't have credentials.
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

func DocketCountFromCaseId(baseURL string, refererURL string, client *http.Client, id string) (int, error) {
	var docketCount int
	u, err := url.Parse(baseURL)
	if err != nil {
		return docketCount, err
	}
	q := u.Query()
	q.Set("search", "caseInfo")
	q.Set("caseid", id)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return docketCount, err
	}

	req.Header.Set("User-Agent", "loganamcnichols")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Referer", refererURL)

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

func GetDownloadLink(client *http.Client, url string, referer string, docNo string, caseNum string) (string, string, error) {
	var downloadLink string
	var deSeqNum string
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	field1, err := writer.CreateFormField(fmt.Sprintf("CaseNum_%s", caseNum))
	if err != nil {
		return downloadLink, deSeqNum, err
	}
	field1.Write([]byte("on"))

	field2, err := writer.CreateFormField("document_number")
	if err != nil {
		return downloadLink, deSeqNum, err
	}
	field2.Write([]byte(docNo))
	writer.Close()

	req, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return downloadLink, deSeqNum, err
	}
	if err != nil {
		return downloadLink, deSeqNum, err
	}
	req.Header.Set("User-Agent", "loganamcnichols")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Referer", referer)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return downloadLink, deSeqNum, err
	}
	if err != nil {
		return downloadLink, deSeqNum, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return downloadLink, deSeqNum, fmt.Errorf("recieved non found code")
	}

	urlObj := resp.Request.URL
	deSeqNum = urlObj.Query().Get("de_seq_num")
	urlObj.RawQuery = ""
	downloadLink = urlObj.String()

	return downloadLink, deSeqNum, nil

}

func AppendToEnvFile(key, value string) error {
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

func GetCaseMainPage(client *http.Client, url string, case_id string, case_number string) (*goquery.Document, error) {
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
	field2.Write([]byte(case_id))
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

func GetFormURL(client *http.Client, queryURL string) (string, error) {
	var caseURL string
	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return caseURL, err
	}
	req.Header.Set("User-Agent", "loganamcnichols")
	req.Header.Set("Accept", "text/html")

	resp, err := client.Do(req)
	if err != nil {
		return caseURL, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return caseURL, fmt.Errorf("recieved non found code")
	}
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return caseURL, err
	}
	caseAction, exists := document.Find("form").First().Attr("action")
	if !exists {
		return caseURL, fmt.Errorf("no action attribute found")
	}
	baseURL, err := url.Parse(queryURL)
	if err != nil {
		return caseURL, err
	}
	actionURL, err := url.Parse(caseAction)
	if err != nil {
		return caseURL, err
	}
	caseURL = baseURL.ResolveReference(actionURL).String()
	return caseURL, nil
}

func PurchaseDocument(client *http.Client, reqURL string, caseID string, deSeqNum string) (*goquery.Document, error) {
	var document *goquery.Document
	buffer := bytes.Buffer{}
	writer := multipart.NewWriter(&buffer)

	field1, err := writer.CreateFormField("caseid")
	if err != nil {
		return document, err
	}
	field1.Write([]byte(caseID))
	field2, err := writer.CreateFormField("de_seq_num")
	if err != nil {
		return document, err
	}
	field2.Write([]byte(deSeqNum))
	field3, err := writer.CreateFormField("got_receipt")
	if err != nil {
		return document, err
	}
	field3.Write([]byte("1"))
	writer.Close()
	field4, err := writer.CreateFormField("pd_toggle_possible")
	if err != nil {
		return document, err
	}
	field4.Write([]byte("1"))
	writer.Close()

	req, err := http.NewRequest("POST", reqURL, &buffer)
	if err != nil {
		return document, err
	}

	req.Header.Set("User-Agent", "loganamcnichols")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Referer", reqURL)
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

func PerformDownload(client *http.Client, doc *goquery.Document, baseURL string, caseID string, docNum string) error {
	var err error
	src, exists := doc.Find("iframe").First().Attr("src")
	if !exists {
		return fmt.Errorf("no src attribute found")
	}
	baseURLObj, err := url.Parse(baseURL)
	if err != nil {
		return err
	}
	pdfURLObj, err := url.Parse(src)
	if err != nil {
		return err
	}
	fullURLObj := baseURLObj.ResolveReference(pdfURLObj)
	fullURL := fullURLObj.String()
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "loganamcnichols")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Referer", baseURL)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("recieved non found code")
	}
	dest := fmt.Sprintf("pdfs/%s-%s.pdf", caseID, docNum)
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func GetDocumentSummary(client *http.Client, url string, caseID string) (*goquery.Document, error) {
	var document *goquery.Document
	buffer := bytes.Buffer{}
	writer := multipart.NewWriter(&buffer)

	field1, err := writer.CreateFormField("view_comb_doc_text")
	if err != nil {
		return document, err
	}
	field1.Write([]byte(""))

	field2, err := writer.CreateFormField("all_case_ids")
	if err != nil {
		return document, err
	}
	field2.Write([]byte(caseID))
	field3, err := writer.CreateFormField(fmt.Sprintf("CaseNum_%s", caseID))
	if err != nil {
		return document, err
	}
	field3.Write([]byte("on"))
	field4, err := writer.CreateFormField("date_from")
	if err != nil {
		return document, err
	}
	field4.Write([]byte(""))
	field5, err := writer.CreateFormField("date_range_type")
	if err != nil {
		return document, err
	}
	field5.Write([]byte("Filed"))
	field6, err := writer.CreateFormField("date_from")
	if err != nil {
		return document, err
	}
	field6.Write([]byte(""))
	field7, err := writer.CreateFormField("date_to")
	if err != nil {
		return document, err
	}
	field7.Write([]byte(""))
	field8, err := writer.CreateFormField("documents_numbered_from_")
	if err != nil {
		return document, err
	}
	field8.Write([]byte(""))
	field9, err := writer.CreateFormField("list_of_parties_and_counsel")
	if err != nil {
		return document, err
	}
	field9.Write([]byte("on"))
	field10, err := writer.CreateFormField("terminated_parties")
	if err != nil {
		return document, err
	}
	field10.Write([]byte("on"))
	field11, err := writer.CreateFormField("pdf_header")
	if err != nil {
		return document, err
	}
	field11.Write([]byte("pdf_header"))
	field12, err := writer.CreateFormField("output_format")
	if err != nil {
		return document, err
	}
	field12.Write([]byte("hml"))
	field13, err := writer.CreateFormField("PreResetField")
	if err != nil {
		return document, err
	}
	field13.Write([]byte(""))
	field14, err := writer.CreateFormField("PreResetFields")
	if err != nil {
		return document, err
	}
	field14.Write([]byte(""))
	field15, err := writer.CreateFormField("sort1")
	if err != nil {
		return document, err
	}
	field15.Write([]byte("oldest date first"))
	writer.Close()

	req, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return document, err
	}
	req.Header.Set("User-Agent", "loganamcnichols")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Referer", url)
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
