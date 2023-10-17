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

type BrowseDocsTemplate struct {
	UserID        int
	PacerLoggedIn bool
	Docs          []BrowseDocs
	Credits       int
}

func BrowseHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/browse.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	userID := CheckSession(r)

	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}

	creditRows := cnx.QueryRow("SELECT credits FROM users WHERE id = $1", userID)
	var credits int
	err = creditRows.Scan(&credits)
	if err != nil {
		log.Printf("Error scanning row: %v", err)
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

	data := BrowseDocsTemplate{
		UserID:        userID,
		PacerLoggedIn: CheckPacerSession(r),
		Docs:          docs,
		Credits:       credits,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
