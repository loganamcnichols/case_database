package handlers

import (
	"fmt"
	"net/http"

	"github.com/loganamcnichols/case_database/pkg/db"
)

func CreditsHandler(w http.ResponseWriter, r *http.Request) {
	userID := CheckSession(r)
	con, err := db.Connect()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer con.Close()
	var credits int
	err = con.QueryRow("SELECT credits FROM users WHERE id = $1", userID).Scan(&credits)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	creditMsg := fmt.Sprintf("Available Credits: %d", credits)
	w.Write([]byte(creditMsg))
}
