package db

import (
	"testing"
)

func TestConnect(t *testing.T) {
	db, err := Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	defer db.Close()
}
