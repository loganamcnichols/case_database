package handlers

import (
	"bytes"
	"html/template"
	"net/http"
)

type HomeTemplateData struct {
	UserID        int
	PacerLoggedIn bool
	Credits       int
}

func LoadPage(w http.ResponseWriter, r *http.Request, page string, data any) {
	var mainBuf bytes.Buffer
	var headerBuf bytes.Buffer
	if CheckSession(r) > 0 {
		headerTemplate, err := template.ParseFiles("web/templates/member-header.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = headerTemplate.Execute(&headerBuf, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		headerTemplate, err := template.ParseFiles("web/templates/guest-header.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = headerTemplate.Execute(&headerBuf, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	mainTemplate, err := template.ParseFiles(page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = mainTemplate.Execute(&mainBuf, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	layoutTemplate, err := template.ParseFiles("web/templates/layout.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = layoutTemplate.Execute(w, struct {
		Header template.HTML
		Main   template.HTML
	}{
		Header: template.HTML(headerBuf.String()),
		Main:   template.HTML(mainBuf.String()),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Exec full page reload if needed.
	if isHtmx := r.Header.Get("HX-Request"); isHtmx != "true" {
		LoadPage(w, r, "web/templates/home.html", nil)
		return
	}
	tmpl, err := template.ParseFiles("web/templates/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
