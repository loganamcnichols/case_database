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

type HomeTemplateData struct {
	UserID        int
	PacerLoggedIn bool
	Cases         []Case
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/home.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	loggedIn := CheckSession(r)

	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
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

	data := HomeTemplateData{
		UserID:        loggedIn,
		PacerLoggedIn: CheckPacerSession(r),
		Cases:         cases,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
