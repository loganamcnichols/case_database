package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	dbPassword := os.Getenv("DB_PASSWORD")
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

func QueryCases(cnx *sql.DB, courtID string, caseID int) ([]string, error) {
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
