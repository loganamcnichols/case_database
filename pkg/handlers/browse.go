package handlers

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/loganamcnichols/case_database/pkg/db"
)

type Case struct {
	ID      int    `db:"id"`
	PacerID int    `db:"pacer_id"`
	CourtID string `db:"court_id"`
	Title   string `db:"title"`
	Number  string `db:"case_number"`
}

type Doc struct {
	Title       string `db:"title"`
	ID          int    `db:"id"`
	Description string `db:"description"`
	File        string `db:"file"`
	DocNumber   int    `db:"doc_number"`
	CaseID      int    `db:"case_id"`
	Pages       int    `db:"pages"`
	UserID      int    `db:"user_id"`
	PacerID     string `db:"pacer_id"`
	Court       string `db:"court"`
}

type DocInfo struct {
	Doc
	Credits int
}

func BrowseHandler(w http.ResponseWriter, r *http.Request) {
	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}
	rows, err := cnx.Query("SELECT * FROM cases LIMIT 20")
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	var cases []Case
	var c Case
	for rows.Next() {
		if err := rows.Scan(&c.ID, &c.PacerID, &c.CourtID, &c.Title, &c.Number); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue // Skip this iteration and move to the next one
		}
		cases = append(cases, c)
	}
	data := struct {
		Cases  []Case
		Search string
		CaseID int
	}{
		Cases:  cases,
		CaseID: c.ID,
	}
	// Exec full page reload if needed.
	if isHtmx := r.Header.Get("HX-Request"); isHtmx != "true" {
		LoadPage(w, r, "web/templates/browse.html", &data)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/browse.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, &data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func BrowseSearchHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}
	rows, err := cnx.Query("SELECT * FROM cases WHERE title ILIKE '%' || $1 || '%' LIMIT 20", search)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	var cases []Case
	var c Case
	for rows.Next() {
		if err := rows.Scan(&c.ID, &c.PacerID, &c.CourtID, &c.Title, &c.Number); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue // Skip this iteration and move to the next one
		}
		cases = append(cases, c)
	}

	data := struct {
		Cases  []Case
		Search string
		caseID int
	}{
		Cases:  cases,
		Search: search,
		caseID: c.ID,
	}

	// Exec full page reload if needed.
	if isHtmx := r.Header.Get("HX-Request"); isHtmx != "true" {
		LoadPage(w, r, "web/templates/browse-search.html", &data)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/browse-search.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, &data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func BrowseScrollHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	caseID := r.URL.Query().Get("caseID")
	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}
	rows, err := cnx.Query("SELECT * FROM cases WHERE title ILIKE '%' || $1 || '%' AND id > $2 LIMIT 20", search, caseID)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	var cases []Case
	var c Case
	for rows.Next() {
		if err := rows.Scan(&c.ID, &c.PacerID, &c.CourtID, &c.Title, &c.Number); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue // Skip this iteration and move to the next one
		}
		cases = append(cases, c)
	}
	data := struct {
		Cases  []Case
		Search string
		caseID int
	}{
		Cases:  cases,
		caseID: c.ID,
	}
	// Exec full page reload if needed.
	if isHtmx := r.Header.Get("HX-Request"); isHtmx != "true" {
		LoadPage(w, r, "web/templates/browse-scroll.html", &data)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/browse-scroll.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, &data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func BrowseDocsHandler(w http.ResponseWriter, r *http.Request) {
	caseID := r.URL.Query().Get("caseID")
	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}
	rows, err := cnx.Query("SELECT * FROM documents WHERE case_id = $1", caseID)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	var docs []DocInfo
	var d DocInfo
	for rows.Next() {
		if err := rows.Scan(&d.ID, &d.Description, &d.File, &d.DocNumber, &d.CaseID, &d.Pages, &d.UserID, &d.PacerID, &d.Court); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue // Skip this iteration and move to the next one
		}
		d.Credits = d.Pages * 50
		docs = append(docs, d)
	}

	data := struct {
		Docs []DocInfo
	}{
		Docs: docs,
	}
	tmpl, err := template.ParseFiles("web/templates/browse-docs.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, &data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func PurchaseDocCreditsHandler(w http.ResponseWriter, r *http.Request) {
	docID := r.URL.Query().Get("docID")
	file := r.URL.Query().Get("file")
	creditsDue, err := strconv.Atoi(r.URL.Query().Get("credits"))
	if err != nil {
		log.Println(err)
	}

	userID := CheckSession(r)
	if userID == 0 {
		w.Header().Set("HX-Retarget", "main")
		LoginHandler(w, r)
		return
	}

	cnx, err := db.Connect()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		cnx.Close()
		return
	}
	defer cnx.Close()
	var credits int
	cnx.QueryRow("SELECT credits FROM users WHERE id = $1", userID).Scan(&credits)

	if credits > creditsDue {
		cnx.Exec("INSERT INTO users_by_documents (user_id, doc_id) VALUES ($1, $2)", userID, docID)
		cnx.Exec("UPDATE users SET credits = credits - $1 WHERE id = $2", creditsDue, userID)
	} else {
		w.Write([]byte("Not enough credits"))
	}

	tmpl, err := template.ParseFiles("web/templates/view-pdf.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, "/pdfs/"+file)
}

func ViewPDFHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/pdf")
	filePath := r.URL.Path[1:]
	pathPart := strings.Split(filePath, "/")
	file := pathPart[len(pathPart)-1]
	userID := CheckSession(r)
	if userID == 0 {
		w.Header().Set("HX-Retarget", "main")
		LoginHandler(w, r)
		return
	}
	cnx, err := db.Connect()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		cnx.Close()
		return
	}
	defer cnx.Close()

	rows := cnx.QueryRow("SELECT user_id, doc_id FROM users_by_documents WHERE user_id = $1 AND doc_id IN (SELECT doc_id FROM documents WHERE file = $2)", userID, file)

	err = rows.Scan(&userID, &file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, filePath)
}
