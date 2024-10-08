package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type CheckoutTemplateData struct {
	Title         string
	Amount        string
	Dollars       string
	UserID        int
	PacerLoggedIn bool
}

func CheckoutHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		return
	}
	amount := r.FormValue("amount")
	log.Printf("Received amount: %s", amount)
	tmpl, err := template.ParseFiles("web/templates/checkout.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	items := []Item{
		{
			Id:     "credits",
			Amount: amount,
		},
	}

	cents := CalculateOrderAmount(items)

	dollars := "$" + fmt.Sprintf("%.2f", float64(cents)/100.0)

	data := struct {
		Title         string
		Amount        string
		Dollars       string
		UserID        int
		PacerLoggedIn bool
	}{
		Title:         "Pacer Lookup - Case Database",
		Amount:        amount,
		Dollars:       dollars,
		UserID:        CheckSession(r),
		PacerLoggedIn: CheckPacerSession(r),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}
