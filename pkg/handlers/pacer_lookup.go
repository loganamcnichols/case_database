package handlers

import (
	"html/template"
	"log"
	"net/http"
	"net/url"

	"github.com/loganamcnichols/case_database/pkg/scraper"
)

type PacerLookupTemplateData struct {
	Title         string
	LoggedIn      bool
	PacerLoggedIn bool
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

func PacerLoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/pacer-login.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title         string
		LoggedIn      bool
		PacerLoggedIn bool
	}{
		Title:         "Pacer Login - Case Database",
		LoggedIn:      CheckSession(r),
		PacerLoggedIn: CheckPacerSession(r),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func PacerLoginSubmitHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	log.Printf("Received username: %s", username)
	log.Printf("Received password: %s", password)

	client, err := scraper.LoginToPacer(username, password, "")
	if err != nil {
		log.Printf("Error logging in to Pacer: %s", err)
		http.Error(w, "Error logging in to Pacer", http.StatusInternalServerError)
		return
	}
	u, _ := url.Parse(scraper.LoginURL)
	var nextGenCSO string
	cookies := client.Jar.Cookies(u)
	for _, cookie := range cookies {
		if cookie.Name == "NextGenCSO" {
			nextGenCSO = cookie.Value
			break
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "NextGenCSO",
		Value:    nextGenCSO,
		Path:     "/",
		HttpOnly: true,
		// Add other cookie settings like Secure, SameSite, etc., as needed.
	})
	http.Redirect(w, r, "/pacer-lookup", http.StatusSeeOther) // Redirect to home page or dashboard
}
