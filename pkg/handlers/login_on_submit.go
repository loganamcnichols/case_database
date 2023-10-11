package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sync"

	"github.com/loganamcnichols/case_database/pkg/db"
)

var sessionStore = make(map[string]int) // map[sessionID]userID
var sessionMutex = &sync.RWMutex{}

func generateSessionID() string {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func LoginOnSubmitHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	log.Printf("Received email: %s, password: %s", email, password)

	cnx, err := db.Connect()
	if err != nil {
		log.Println(err)
	}
	defer cnx.Close()
	userID, err := db.GetUserID(cnx, email, password)
	if err != nil {
		http.Error(w, "Error getting user ID", http.StatusInternalServerError)
	}
	// Generate a session ID
	sessionID := generateSessionID()

	// Store the session ID and user ID in the session store
	sessionMutex.Lock()
	sessionStore[sessionID] = userID
	sessionMutex.Unlock()

	// Set the session ID in a cookie for the user
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		// Add other cookie settings like Secure, SameSite, etc., as needed.
	})

	http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect to home page or dashboard
}
