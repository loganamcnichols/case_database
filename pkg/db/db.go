package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type Execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row

	// Add other methods like Prepare if needed
}

func Connect() (*sql.DB, error) {
	dbPassword := os.Getenv("PGPASSWORD")
	connStr := fmt.Sprintf("user=logan dbname=casedatabase password=%s host=localhost sslmode=disable", dbPassword)

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
		fmt.Println("Successfully connected to the database.")
	}
	return db, err
}

func QueryCases(cnx Execer, courtID string, caseID int) ([]string, error) {
	rows, err := cnx.Query("SELECT title FROM cases WHERE court_id=$1 AND pacer_id=$2", courtID, caseID)
	if err != nil {
		log.Fatal("Error querying database:", err)
	}
	defer rows.Close()
	var titles []string
	for rows.Next() {
		var title string
		err := rows.Scan(&title)
		if err != nil {
			log.Fatal("Error scanning rows:", err)
		}
		titles = append(titles, title)
	}
	return titles, err
}

func InsertCases(cnx Execer, courtID string, caseID int, title string) error {
	_, err := cnx.Exec("INSERT INTO cases (court_id, pacer_id, title) VALUES ($1, $2, $3)", courtID, caseID, title)
	if err != nil {
		log.Fatal("Error inserting into database:", err)
	}
	return err
}
