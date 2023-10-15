package handlers

import (
	"html/template"
	"log"
	"net/http"
)

type BuyCreditsTemplateData struct {
	Title  string
	UserID int
}

func BuyCreditsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/buy-credits.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	loggedIn := CheckSession(r)

	data := struct {
		Title         string
		UserID        int
		PacerLoggedIn bool
	}{
		Title:         "Pacer Documents Resale Market",
		UserID:        loggedIn,
		PacerLoggedIn: CheckPacerSession(r),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func BuyCreditsOnSubmit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	amount := r.FormValue("amount")
	log.Printf("Received amount: %s", amount)
}
