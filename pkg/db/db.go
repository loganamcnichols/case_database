package db

import (
	"database/sql"
	"errors"
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

func InsertCases(cnx Execer, courtID string, caseID int, title string, caseNumber string) error {
	_, err := cnx.Exec("INSERT INTO cases (court_id, pacer_id, title, case_number) VALUES ($1, $2, $3, $4)", courtID, caseID, title, caseNumber)
	if err != nil {
		log.Fatal("Error inserting into database:", err)
	}
	return err
}

func Head(cnx Execer) (*sql.Rows, error) {
	rows, err := cnx.Query("SELECT * FROM cases LIMIT 20")
	if err != nil {
		log.Fatal("Error querying database:", err)
	}
	return rows, err
}

func CreateUser(cnx Execer, email string, password string) error {
	rows, err := cnx.Query("SELECT * FROM users WHERE email=$1", email)
	if err != nil {
		log.Fatal("Error querying database:", err)
	}
	if rows.Next() {
		return errors.New("email already in use")
	}
	_, err = cnx.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", email, password)
	if err != nil {
		log.Fatal("Error querying database:", err)
	}
	return err
}

func GetUser(cnx Execer, email string, password string) (*sql.Rows, error) {
	rows, err := cnx.Query("SELECT * FROM users WHERE email=$1 AND password=$2", email, password)
	if err != nil {
		log.Fatal("Error querying database:", err)
	}
	return rows, err
}
