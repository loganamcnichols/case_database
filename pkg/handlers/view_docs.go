package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/loganamcnichols/case_database/pkg/db"
)

type ViewDocsTemplateData struct {
	LoggedIn      bool
	PacerLoggedIn bool
	Docs          []Doc
}

type Doc struct {
	ID          int    `db:"id"`
	Description string `db:"description"`
	File        string `db:"file"`
	DocNumber   int    `db:"doc_number"`
}

func ViewDocsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/view-docs.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		log.Printf("Error getting session cookie: %v", err)
	}

	sessionMutex.RLock()
	userID := sessionStore[sessionID.Value]
	sessionMutex.RUnlock()

	loggedIn := CheckSession(r)

	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
		defer cnx.Close()
	}

	rows, err := db.QueryUserDocs(cnx, userID)
	if err != nil {
		log.Printf("Error getting top rows: %v", err)
	}

	var docs []Doc
	for rows.Next() {
		var d Doc
		err = rows.Scan(&d.ID, &d.Description, &d.File, &d.DocNumber)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
		}
		docs = append(docs, d)
	}

	data := ViewDocsTemplateData{
		LoggedIn:      loggedIn,
		PacerLoggedIn: CheckPacerSession(r),
		Docs:          docs,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
