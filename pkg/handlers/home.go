package handlers

import (
	"html/template"
	"net/http"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/home.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title string
	}{
		Title: "Pacer Documents Resale Market",
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
