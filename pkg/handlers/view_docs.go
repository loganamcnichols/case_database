package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/loganamcnichols/case_database/pkg/db"
)

type ViewDocsTemplateData struct {
	UserID        int
	PacerLoggedIn bool
	Docs          []Doc
	Credits       int
}

func UserBrowseHandler(w http.ResponseWriter, r *http.Request) {
	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}

	userID := CheckSession(r)
	rows, err := cnx.Query("SELECT * FROM cases WHERE pacer_id IN (SELECT case_id FROM documents WHERE id IN (SELECT doc_id FROM users_by_documents WHERE user_id = $1)) LIMIT 20", userID)
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
		LoadPage(w, r, "web/templates/user-browse.html", &data)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/user-browse.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, &data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func UserBrowseSearchHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}
	userID := CheckSession(r)
	rows, err := cnx.Query("SELECT * FROM cases WHERE title ILIKE '%' || $1 || '%' AND pacer_id IN (SELECT case_id FROM documents WHERE id IN (SELECT doc_id FROM users_by_documents WHERE user_id = $2)) LIMIT 20", search, userID)
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
		LoadPage(w, r, "web/templates/user-browse-search.html", &data)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/user-browse-search.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, &data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func UserBrowseScrollHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	caseID := r.URL.Query().Get("caseID")
	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}
	userID := CheckSession(r)
	rows, err := cnx.Query("SELECT * FROM cases WHERE pacer_id IN (SELECT case_id FROM documents WHERE id IN (SELECT doc_id FROM users_by_documents WHERE user_id = $1)) AND title ILIKE '%' || $2 || '%' AND id > $3 LIMIT 20", userID, search, caseID)
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
		LoadPage(w, r, "web/templates/user-browse-scroll.html", &data)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/user-browse-scroll.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, &data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func UserBrowseDocsHandler(w http.ResponseWriter, r *http.Request) {
	caseID := r.URL.Query().Get("caseID")
	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}
	userID := CheckSession(r)
	rows, err := cnx.Query("SELECT * FROM documents WHERE case_id = $1 AND id IN (SELECT doc_id FROM users_by_documents WHERE user_id = $2)", caseID, userID)
	if err != nil {
		log.Println(err)
	}
	defer rows.Close()
	var docs []DocInfo
	var d Doc
	for rows.Next() {
		if err := rows.Scan(&d.ID, &d.Description, &d.File, &d.DocNumber, &d.CaseID, &d.Pages, &d.UserID, &d.PacerID, &d.Court, &d.StartDate); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue // Skip this iteration and move to the next one
		}
		var credits int64
		if d.Pages.Valid {
			credits = d.Pages.Int64 * 50
		}
		docs = append(docs, DocInfo{
			d.File.String,
			d.ID,
			credits,
			d.Description.String,
			d.DocNumber,
			d.Pages.Int64,
		})
	}

	data := struct {
		Docs    []DocInfo
		PacerID string
	}{
		Docs:    docs,
		PacerID: caseID,
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
