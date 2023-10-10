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
