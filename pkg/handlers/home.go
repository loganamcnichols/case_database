package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/loganamcnichols/case_database/pkg/db"
)

type HomeTemplateData struct {
	UserID        int
	PacerLoggedIn bool
	Credits       int
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/home.html")
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

	data := HomeTemplateData{
		UserID:        userID,
		PacerLoggedIn: CheckPacerSession(r),
		Credits:       credits,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
