package handlers

import (
	"html/template"
	"net/http"
)

type LoginTemplateData struct {
	Title    string
	LoggedIn bool
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/login.html")
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

func CheckSession(r *http.Request) bool {
	// Check if the user has a session ID cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return false
	}

	// Check if the session ID is in the session store
	sessionMutex.RLock()
	_, ok := sessionStore[cookie.Value]
	sessionMutex.RUnlock()
	return ok
}
