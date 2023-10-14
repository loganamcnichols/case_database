package handlers

import (
	"html/template"
	"net/http"
)

type PacerLookupTemplateData struct {
	Title    string
	LoggedIn bool
}

func PacerLookupHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("NextGenCSO"); err != nil {
		http.Redirect(w, r, "/pacer-login", http.StatusTemporaryRedirect)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/pacer-lookup.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title    string
		LoggedIn bool
	}{
		Title:    "Pacer Lookup - Case Database",
		LoggedIn: CheckSession(r),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func PacerLoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/pacer-login.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title    string
		LoggedIn bool
	}{
		Title:    "Pacer Login - Case Database",
		LoggedIn: CheckSession(r),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
