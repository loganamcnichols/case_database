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

type BrowseTemplateData struct {
	UserID        int
	PacerLoggedIn bool
	Cases         []Case
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

	rows, err := db.Head(cnx)
	if err != nil {
		log.Printf("Error getting top rows: %v", err)
	}

	var cases []Case
	for rows.Next() {
		var c Case
		err = rows.Scan(&c.ID, &c.PacerID, &c.CourtID, &c.Title, &c.Number)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
		}
		cases = append(cases, c)
	}

	data := BrowseTemplateData{
		UserID:        userID,
		PacerLoggedIn: CheckPacerSession(r),
		Cases:         cases,
		Credits:       credits,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
