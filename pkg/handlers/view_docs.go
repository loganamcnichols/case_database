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

type Doc struct {
	Title       string `db:"title"`
	ID          int    `db:"id"`
	Description string `db:"description"`
	File        string `db:"file"`
	DocNumber   int    `db:"doc_number"`
	CaseID      int    `db:"case_id"`
}

func ViewDocsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/view-docs.html")
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

	rows, err := db.QueryUserDocs(cnx, userID)
	if err != nil {
		log.Printf("Error getting top rows: %v", err)
	}

	var docs []Doc
	for rows.Next() {
		var d Doc
		err = rows.Scan(&d.Title, &d.ID, &d.Description, &d.File, &d.DocNumber, &d.CaseID)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
		}
		docs = append(docs, d)
	}

	data := ViewDocsTemplateData{
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
