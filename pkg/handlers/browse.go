package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/loganamcnichols/case_database/pkg/db"
)

type Case struct {
	ID      int    `db:"id"`
	PacerID int    `db:"pacer_id"`
	CourtID string `db:"court_id"`
	Title   string `db:"title"`
	Number  string `db:"case_number"`
}

type BrowseDocs struct {
	Title       string `db:"title"`
	ID          int    `db:"id"`
	Description string `db:"description"`
	File        string `db:"file"`
	DocNumber   int    `db:"doc_number"`
	CaseID      int    `db:"case_id"`
	Pages       int    `db:"pages"`
	UserID      int    `db:"user_id"`
	Cost        int
}

func BrowseHandler(w http.ResponseWriter, r *http.Request) {

	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}

	rows, err := db.QueryDocs(cnx)
	if err != nil {
		log.Printf("Error getting top rows: %v", err)
	}

	var docs []BrowseDocs
	for rows.Next() {
		var d BrowseDocs
		err = rows.Scan(&d.Title, &d.ID, &d.Description, &d.File, &d.DocNumber, &d.CaseID, &d.Pages, &d.UserID)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
		}
		d.Cost = d.Pages * 10
		docs = append(docs, d)
	}

	// Exec full page reload if needed.
	if isHtmx := r.Header.Get("HX-Request"); isHtmx != "true" {
		LoadPage(w, r, "web/templates/browse.html", &docs)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/browse.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, &docs)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
