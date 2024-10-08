package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
)

func ConnectTest() (*sql.DB, error) {
	dbPassword := os.Getenv("PGPASSWORD")
	connStr := fmt.Sprintf("user=logan dbname=test_db password=%s host=localhost sslmode=disable", dbPassword)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()

	// Your database operations here
	err = db.Ping()
	if err != nil {
		log.Fatal("Error while pinging database:", err)
	} else {
		fmt.Println("Successfully ConnectTested to the database.")
	}
	return db, err
}

func TestQueryCases(t *testing.T) {
	db, err := ConnectTest()
	if err != nil {
		t.Errorf("Error ConnectTesting to database: %v", err)
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
	db, err := ConnectTest()
	if err != nil {
		t.Errorf("Error ConnectTesting to database: %v", err)
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		t.Errorf("Error beginning transaction: %v", err)
	}
	err = InsertCases(tx, "azd", 1303801, "Test Case", "2:22-mj-2189")
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
	db, err := ConnectTest()
	if err != nil {
		t.Errorf("Error ConnectTesting to database: %v", err)
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

func TestUpdateUserCredits(t *testing.T) {
	con, err := ConnectTest()
	if err != nil {
		t.Errorf("Error ConnectTesting to database: %v", err)
	}
	defer con.Close()
	UpdateUserCredits(con, 1, 1000)
	row := con.QueryRow("SELECT credits FROM users WHERE id = 1")
	var credits int
	err = row.Scan(&credits)
	if err != nil {
		t.Errorf("Error scanning row: %v", err)
	}
	if credits != 1000 {
		t.Errorf("updateUserCredits() did not update credits")
	}
	UpdateUserCredits(con, 1, 0)
}
