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

func TestQueryUserDocs(t *testing.T) {
	db, err := Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	defer db.Close()
	rows, err := QueryUserDocs(db, 5)
	if err != nil {
		t.Errorf("Error querying user docs: %v", err)
	}
	var id int
	var caseTitle string
	var descritpion string
	var file string
	var docNumber string
	var caseID int
	rows.Next()
	if err := rows.Scan(&caseTitle, &id, &descritpion, &file, &docNumber, &caseID); err != nil {
		t.Errorf("Error scanning rows: %v", err)
	}
	if caseTitle != "2:22-cv-02189-SRB Stanley v. Quintairos Prieto Wood & Boyer PA" {
		t.Errorf("QueryUserDocs() returned wrong caseTitle")
	}
	if descritpion != "CORPORATE DISCLOSURE STATEMENT" {
		t.Errorf("QueryUserDocs() returned wrong description")
	}
	if file != "1320666-2.pdf" {
		t.Errorf("QueryUserDocs() returned wrong file")
	}
	if docNumber != "2" {
		t.Errorf("QueryUserDocs() returned wrong docNumber")
	}
}

func TestUpdateUserCredits(t *testing.T) {
	con, err := Connect()
	if err != nil {
		t.Errorf("Error connecting to database: %v", err)
	}
	defer con.Close()
	UpdateUserCredits(con, 5, 1000)
	row := con.QueryRow("SELECT credits FROM users WHERE id = 5")
	var credits int
	err = row.Scan(&credits)
	if err != nil {
		t.Errorf("Error scanning row: %v", err)
	}
	if credits != 1000 {
		t.Errorf("updateUserCredits() did not update credits")
	}
	UpdateUserCredits(con, 5, 0)
}
