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
	err = InsertCases(db, "azd", 1303801, "Test Case")
	if err != nil {
		t.Errorf("Error inserting cases: %v", err)
	}
	res, err := QueryCases(db, "azd", 1303801)
	if err != nil {
		t.Errorf("Error querying database: %v", err)
	}
	if len(res) == 0 {
		t.Errorf("InsertCases() did not insert case")
	}

	db.Exec("DELETE FROM cases WHERE court_id='azd' AND pacer_id=1303801")
}
