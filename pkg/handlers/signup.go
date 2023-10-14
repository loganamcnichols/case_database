package handlers

import (
	"html/template"
	"net/http"
)

type SignupTemplateData struct {
	Title         string
	LoggedIn      bool
	PacerLoggedIn bool
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/signup.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title         string
		LoggedIn      bool
		PacerLoggedIn bool
	}{
		Title:         "Pacer Lookup - Case Database",
		LoggedIn:      CheckSession(r),
		PacerLoggedIn: CheckPacerSession(r),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
