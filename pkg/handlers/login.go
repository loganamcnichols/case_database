package handlers

import (
	"html/template"
	"net/http"
)

type LoginTemplateData struct {
	Title         string
	UserID        int
	PacerLoggedIn bool
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("web/templates/login.html")
	if err != nil {
		http.Error(w, "Could not load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title         string
		UserID        int
		PacerLoggedIn bool
	}{
		Title:         "Pacer Lookup - Case Database",
		UserID:        CheckSession(r),
		PacerLoggedIn: CheckPacerSession(r),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Could not write template", http.StatusInternalServerError)
	}
}

func CheckSession(r *http.Request) int {
	// Check if the user has a session ID cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		return 0
	}

	// Check if the session ID is in the session store
	sessionMutex.RLock()
	val := sessionStore[cookie.Value]
	sessionMutex.RUnlock()
	return val
}

func CheckPacerSession(r *http.Request) bool {
	// Check if the user has a nextgencsocookie
	_, err := r.Cookie("NextGenCSO")
	return err == nil
}
