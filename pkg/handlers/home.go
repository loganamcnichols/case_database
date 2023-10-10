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
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/home.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

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
		err = rows.Scan(&c.ID, &c.PacerID, &c.CourtID, &c.Title)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
		}
		cases = append(cases, c)
	}

	err = tmpl.Execute(w, cases)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
