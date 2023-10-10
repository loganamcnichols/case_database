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
