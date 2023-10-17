package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

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
	// Hashing the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = cnx.Exec("INSERT INTO users (email, password, credits) VALUES ($1, $2, 0)", email, hashedPassword)
	if err != nil {
		return err
	}
	return err
}

func GetUserID(cnx Execer, email string, password string) (int, error) {
	// Then you verify the provided plainPassword against the stored hashedPassword

	user := cnx.QueryRow("SELECT * FROM users WHERE email=$1", email)
	var hashedPassword []byte
	var credits int64
	var userID int
	err := user.Scan(&userID, &email, &hashedPassword, &credits)
	if err != nil {
		return userID, errors.New("user profile not found")
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		// Passwords do not match or another error occurred
		return userID, errors.New("passwords do not match")
	}
	return userID, err
}

func QueryUserDocs(cnx Execer, userID int) (*sql.Rows, error) {
	rows, err := cnx.Query(`
	SELECT cases.title, documents.* FROM documents
	INNER JOIN users_by_documents ON documents.id = users_by_documents.doc_id 
	INNER JOIN cases ON documents.case_id = cases.pacer_id
	WHERE users_by_documents.user_id = $1`, userID)
	return rows, err
}

func UpdateUserCredits(cnx Execer, userID int, credits int64) error {
	val, err := cnx.Exec("UPDATE users SET credits = $1 WHERE id = $2", credits, userID)
	fmt.Println(val.RowsAffected())
	if err != nil {
		log.Fatal("Error updating user credits:", err)
	}
	return err
}
