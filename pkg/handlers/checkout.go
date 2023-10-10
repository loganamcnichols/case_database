package handlers

import (
	"html/template"
	"log"
	"net/http"
)

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	amount := r.FormValue("amount")
	log.Printf("Received amount: %s", amount)
	tmpl, err := template.ParseFiles("web/templates/checkout.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title  string
		Amount string
	}{
		Title:  "Pacer Lookup - Case Database",
		Amount: amount,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
