package handlers

import (
	"html/template"
	"log"
	"net/http"

	"github.com/loganamcnichols/case_database/pkg/db"
)

type SignupTemplateData struct {
	Title         string
	UserID        int
	PacerLoggedIn bool
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if isHtmx := r.Header.Get("HX-Request"); isHtmx != "true" {
		LoadPage(w, r, "web/templates/signup.html", nil)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/signup.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func SignupOnSubmitHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	log.Printf("Received email: %s, password: %s", email, password)

	cnx, err := db.Connect()
	if err != nil {
		http.Error(w, "Could not connect to database", http.StatusInternalServerError)
		return
	}
	defer cnx.Close()

	err = db.CreateUser(cnx, email, password)
	if err != nil {
		http.Error(w, "Could not create user", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)

}
