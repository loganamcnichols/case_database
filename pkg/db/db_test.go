package db

import (
	"database/sql"
	"testing"
)

func TestConnect(t *testing.T) {
	db, err := Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	defer db.Close()
}

func TestQueryCases(t *testing.T) {
	db, err := Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	defer db.Close()
	cases, err := QueryCases(db, "azd", 1320666)
	if err != nil {
		t.Errorf("Error querying cases: %v", err)
	}
	if len(cases) == 0 {
		t.Errorf("QueryCases() returned no cases")
	}
}

func TestInsertCases(t *testing.T) {
	db, err := Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Errorf("Error beginning transaction: %v", err)
	}
	err = InsertCases(db, "azd", 1303801, "Test Case", "2:22-mj-2189")
	if err != nil {
		t.Errorf("Error inserting cases: %v", err)
	}
	res, err := QueryCases(tx, "azd", 1303801)
	if err != nil {
		t.Errorf("Error querying database: %v", err)
	}
	if len(res) == 0 {
		t.Errorf("InsertCases() did not insert case")
	}
	defer func() {
		err := tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}
	}()
}

func TestUser(t *testing.T) {
	db, err := Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	defer db.Close()
	email := "testuser@test.com"
	password := "testpassword"
	tx, err := db.Begin()
	if err != nil {
		t.Errorf("Error beginning transaction: %v", err)
	}
	err = CreateUser(tx, email, password)
	if err != nil {
		t.Errorf("Error creating user: %v", err)
	}
	res, err := GetUserID(tx, email, password)
	if err != nil {
		t.Errorf("Error getting user: %v", err)
	}
	if res == 0 {
		t.Errorf("GetUser() returned no rows")
	}
	defer func() {
		err := tx.Rollback()
		if err != nil && err != sql.ErrTxDone {
			t.Fatalf("Failed to rollback transaction: %v", err)
		}
	}()
}
