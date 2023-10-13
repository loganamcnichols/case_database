package handlers

import (
	"log"
	"net/http"

	"github.com/loganamcnichols/case_database/pkg/db"
)

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
