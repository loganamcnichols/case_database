package handlers

import (
	"html/template"
	"net/http"
)

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/signup.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
	}{
		Title: "Pacer Lookup - Case Database",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
