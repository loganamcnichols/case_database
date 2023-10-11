package handlers

import (
	"log"
	"net/http"
)

func LoginOnSubmitHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	log.Printf("Received email: %s, password: %s", email, password)
}
