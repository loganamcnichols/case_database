package handlers

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"fmt"
	"regexp"
	"strconv"

	"github.com/loganamcnichols/case_database/pkg/db"
	"github.com/loganamcnichols/case_database/pkg/scraper"
)

type PacerLookupTemplateData struct {
	Title         string
	UserID        int
	PacerLoggedIn bool
}

type DocumentTemplateData struct {
	Count  int
	CaseID string
	Court  string
}

type CaseTemplateData struct {
	Court string
	Cases scraper.PossibleCases
}

func PacerLookupHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("NextGenCSO"); err != nil {
		http.Redirect(w, r, "/pacer-login", http.StatusTemporaryRedirect)
		return
	}
	if isHtmx := r.Header.Get("HX-Request"); isHtmx != "true" {
		LoadPage(w, r, "web/templates/pacer-lookup.html", nil)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/pacer-lookup.html")

	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func PacerLoginHandler(w http.ResponseWriter, r *http.Request) {
	if isHtmx := r.Header.Get("HX-Request"); isHtmx != "true" {
		LoadPage(w, r, "web/templates/pacer-login.html", nil)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/pacer-login.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func PacerLoginSubmitHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	log.Printf("Received username: %s", username)
	log.Printf("Received password: %s", password)

	client, err := scraper.LoginToPacer(username, password, "")
	if err != nil {
		log.Printf("Error logging in to Pacer: %s", err)
		http.Error(w, "Error logging in to Pacer", http.StatusInternalServerError)
		return
	}
	u, _ := url.Parse(scraper.LoginURL)
	var nextGenCSO string
	cookies := client.Jar.Cookies(u)
	for _, cookie := range cookies {
		if cookie.Name == "NextGenCSO" {
			nextGenCSO = cookie.Value
			break
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "NextGenCSO",
		Value:    nextGenCSO,
		Path:     "/",
		HttpOnly: true,
		// Add other cookie settings like Secure, SameSite, etc., as needed.
	})

	tmpl, err := template.ParseFiles("web/templates/pacer-lookup.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	tmpl.Execute(w, nil)
}

func PacerLookupOnSubmit(w http.ResponseWriter, r *http.Request) {
	var possibleCasesTemplate = template.Must(template.ParseFiles("web/templates/possible-cases.html"))
	r.ParseForm()
	court := r.FormValue("court")
	docket := r.FormValue("docket")
	match, _ := regexp.MatchString(`\d{2}-\d{5}`, docket)
	if !match {
		log.Printf("Error verifying docket format")
		http.Error(w, "Error verifying docket format", http.StatusInternalServerError)
		return
	}

	log.Printf("Received court: %s, and docket number: %s", court, docket)
	url := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/possible_case_numbers.pl?%s", court, docket)

	nextGenCSO, _ := r.Cookie("NextGenCSO")
	client, err := scraper.LoginToPacer("", "", nextGenCSO.Value)
	if err != nil {
		w.Header().Set("HX-Retarget", "main")
		PacerLoginHandler(w, r)
		return
	}
	res, err := scraper.PossbleCasesSearch(client, url)
	if err != nil {
		log.Printf("Error searching for possible cases: %v", err)
		http.Error(w, "Error searching for possible cases", http.StatusInternalServerError)
		return
	}

	if len(res.Cases) == 0 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<div>No cases found</div>`))
		return
	}

	templateData := CaseTemplateData{
		Court: court,
		Cases: res,
	}
	possibleCasesTemplate.Execute(w, templateData)
	cnx, err := db.Connect()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		return
	}
	defer cnx.Close()
	for _, c := range res.Cases {
		log.Printf("Case ID: %s, Title: %s", c.ID, c.Title)
		caseID, err := strconv.Atoi(c.ID)
		if err != nil {
			log.Printf("Error converting case ID to int: %v", err)
			return
		}
		err = db.InsertCases(cnx, court, caseID, c.Title, c.Number)
		if err != nil {
			log.Printf("Error inserting case into database: %v", err)
			return
		}
	}
}

func PacerLookupCase(w http.ResponseWriter, r *http.Request) {
	var docketNumber = template.Must(template.ParseFiles("web/templates/docket-number.html"))
	// For hx-get with hx-vals, values are sent as query parameters
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	caseID := r.URL.Query().Get("caseID")
	court := r.URL.Query().Get("court")
	caseNumber := r.URL.Query().Get("caseNumber")

	fmt.Printf("Received caseID: %s, caseNumber: %s\n, courtID: %s\n", caseID, caseNumber, court)

	nextGenCSO, _ := r.Cookie("NextGenCSO")
	client, err := scraper.LoginToPacer("", "", nextGenCSO.Value)
	if err != nil {
		http.Error(w, "Error logging in to PACER", http.StatusInternalServerError)
		return
	}
	baseURL := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/iquery.pl", court)
	moblileURL := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/mobile_query.pl", court)
	caseURL, err := scraper.GetFormURL(client, baseURL)
	if err != nil {
		http.Error(w, "Error getting case URL", http.StatusInternalServerError)
		return
	}
	casePage, err := scraper.GetCaseMainPage(client, caseURL, caseID, caseNumber)
	if err != nil {
		http.Error(w, "Error getting case page", http.StatusInternalServerError)
		return
	}

	metadata := casePage.Find("center")

	metadataHTML, err := metadata.Html()
	if err != nil {
		http.Error(w, "Error getting case metadata", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, metadataHTML)

	count, err := scraper.DocketCountFromCaseId(moblileURL, caseURL, client, caseID)
	if err != nil {
		http.Error(w, "Error getting docket count", http.StatusInternalServerError)
		return
	}
	templateData := DocumentTemplateData{
		Count:  count,
		CaseID: caseID,
		Court:  court,
	}
	fmt.Fprintf(w, "Docket count: %d\n", count)
	docketNumber.Execute(w, templateData)

	con, err := db.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}

	rows, err := con.Query("SELECT doc_number, description FROM documents WHERE case_id = $1", id)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var docNumber int
		var description string
		err = rows.Scan(&docNumber, &description)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Fprintf(w, "DocNumber: %d, Description: %s\n", docNumber, description)
	}

}

func PacerLookupDocketRequest(w http.ResponseWriter, r *http.Request) {

	r.ParseForm()
	docketNumber := r.FormValue("docket-number")
	caseID := r.FormValue("case-id")
	court := r.FormValue("court")

	nextGenCSO, _ := r.Cookie("NextGenCSO")
	// userID := CheckSession(r)

	client, err := scraper.LoginToPacer("", "", nextGenCSO.Value)
	if err != nil {
		log.Printf("Error logging in to PACER: %v", err)
		http.Error(w, "Error logging in to PACER", http.StatusInternalServerError)
		return
	}

	requestURL := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/qryDocument.pl?%s", court, caseID)

	respURL, err := scraper.GetFormURL(client, requestURL)
	if err != nil {
		log.Printf("Error getting form URL: %v", err)
		http.Error(w, "Error getting form URL", http.StatusInternalServerError)
		return
	}
	docIDs, deSeqNum, err := scraper.GetDocIDs(client, respURL, requestURL, docketNumber, caseID)
	downloadLink := fmt.Sprintf("https://ecf.%s.uscourts.gov/doc1/%s", court, docIDs[0])
	if err != nil {
		log.Printf("Error getting download link: %v", err)
		http.Error(w, "Error getting download link", http.StatusInternalServerError)
		return
	}
	pageCount, err := scraper.GetPageCount(client, downloadLink, respURL)
	if err != nil {
		log.Printf("Error getting page count: %v", err)
		http.Error(w, "Error getting page count", http.StatusInternalServerError)
		return
	}
	cost := float32(pageCount) / 10.0
	data := struct {
		DocID        string
		Court        string
		Pages        int
		Cost         string
		CaseID       string
		DeSeqNum     string
		DocketNumber string
	}{
		DocID:        docIDs[0],
		Court:        court,
		Pages:        pageCount,
		Cost:         fmt.Sprintf("$%.2f", cost),
		CaseID:       caseID,
		DeSeqNum:     deSeqNum,
		DocketNumber: docketNumber,
	}
	tmpl, err := template.ParseFiles("web/templates/doc-purchase.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}

	// log.Printf("Received docket number: %s, caseID: %s, court: %s", docketNumber, caseID, court)
	// resDoc, err := scraper.PurchaseDocument(client, downloadLink[0], caseID, deSeqNum)
	// fmt.Println(resDoc.Find("body").Text())
	// if err != nil {
	// 	log.Printf("Error purchasing document: %v", err)
	// 	http.Error(w, "Error purchasing document", http.StatusInternalServerError)
	// 	return
	// }
	// file, err := scraper.PerformDownload(client, resDoc, downloadLink[0], caseID, docketNumber)
	// if err != nil {
	// 	log.Printf("Error performing download: %v", err)
	// 	http.Error(w, "Error performing download", http.StatusInternalServerError)
	// 	return
	// }
	// w.Write([]byte("Downloaded file: " + file + "\n"))

	// cnx, err := db.Connect()
	// if err != nil {
	// 	log.Printf("Error connecting to database: %v", err)
	// 	return
	// }
	// defer cnx.Close()
	// pages, err := scraper.GetPageCount(client, downloadLink[0], respURL)
	// if err != nil {
	// 	log.Printf("Error getting page count: %v", err)
	// 	http.Error(w, "Error getting page count", http.StatusInternalServerError)
	// 	return
	// }

	// var docID int
	// err = cnx.QueryRow(`
	// 	INSERT INTO documents (description, file, doc_number, case_id, pages, user_id)
	// 	VALUES ('description', $1, $2, $3, $4, $5) RETURNING id`,
	// 	file, docketNumber, caseID, pages, userID).Scan(&docID)
	// if err != nil {
	// 	log.Printf("Error inserting into database: %v", err)
	// 	return
	// }
	// _, err = cnx.Exec(`INSERT INTO users_by_documents (user_id, doc_id) VALUES ($1, $2)`, userID, docID)
	// if err != nil {
	// 	log.Printf("Error inserting into database: %v", err)
	// 	return
	// }
}

func PurchaseDocHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	docID := r.FormValue("docID")
	court := r.FormValue("court")
	caseID := r.FormValue("caseID")
	deSeqNum := r.FormValue("deSeqNum")
	docketNumber := r.FormValue("docketNumber")
	pages := r.FormValue("pages")

	nextGenCSO, _ := r.Cookie("NextGenCSO")
	userID := CheckSession(r)

	client, err := scraper.LoginToPacer("", "", nextGenCSO.Value)
	if err != nil {
		log.Printf("Error logging in to PACER: %v", err)
		http.Error(w, "Error logging in to PACER", http.StatusInternalServerError)
		return
	}

	downloadLink := fmt.Sprintf("https://ecf.%s.uscourts.gov/doc1/%s", court, docID)
	resDoc, err := scraper.PurchaseDocument(client, downloadLink, caseID, deSeqNum)
	fmt.Println(resDoc.Find("body").Text())
	if err != nil {
		log.Printf("Error purchasing document: %v", err)
		http.Error(w, "Error purchasing document", http.StatusInternalServerError)
		return
	}
	file, err := scraper.PerformDownload(client, resDoc, downloadLink, caseID, docketNumber)
	if err != nil {
		log.Printf("Error performing download: %v", err)
		http.Error(w, "Error performing download", http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Downloaded file: " + file + "\n"))

	cnx, err := db.Connect()
	if err != nil {
		log.Printf("Error connecting to database: %v", err)
		return
	}
	defer cnx.Close()

	var id int
	err = cnx.QueryRow(`
		INSERT INTO documents (description, file, doc_number, case_id, pages, user_id)
		VALUES ('description', $1, $2, $3, $4, $5) RETURNING id`,
		file, docketNumber, caseID, pages, userID).Scan(&id)
	if err != nil {
		log.Printf("Error inserting into database: %v", err)
		return
	}
	_, err = cnx.Exec(`INSERT INTO users_by_documents (user_id, doc_id) VALUES ($1, $2)`, userID, docID)
	if err != nil {
		log.Printf("Error inserting into database: %v", err)
		return
	}
}

func PacerLookupSummaryRequest(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	caseID := r.FormValue("case-id")
	court := r.FormValue("court")

	nextGenCSO, _ := r.Cookie("NextGenCSO")

	client, err := scraper.LoginToPacer("", "", nextGenCSO.Value)
	if err != nil {
		log.Printf("Error logging in to PACER: %v", err)
		http.Error(w, "Error logging in to PACER", http.StatusInternalServerError)
		return
	}

	requestURL := fmt.Sprintf("https://ecf.%s.uscourts.gov/cgi-bin/DktRpt.pl?%s", court, caseID)
	respURL, err := scraper.GetFormURL(client, requestURL)
	if err != nil {
		log.Printf("Error getting form URL: %v", err)
		http.Error(w, "Error getting form URL", http.StatusInternalServerError)
		return
	}
	document, err := scraper.GetDocumentSummary(client, respURL, caseID)
	if err != nil {
		log.Printf("Error getting document summary: %v", err)
		http.Error(w, "Error getting document summary", http.StatusInternalServerError)
		return
	}

	data, err := document.Find("#cmecfMainContent").First().Html()
	if err != nil {
		log.Printf("Error getting document summary: %v", err)
		http.Error(w, "Error getting document summary", http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, data)
}
