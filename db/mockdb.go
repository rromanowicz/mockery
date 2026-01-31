// Package db
package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB() *sql.DB {
	log.Println("Starting DB")

	DB, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Fatal(err.Error())
	}
	sqlStmt := `
 CREATE TABLE IF NOT EXISTS mock (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  method TEXT,
  path TEXT,
	request_header_matchers TEXT,
	request_query_matchers TEXT,
	request_body_matchers TEXT,
	response_status INTEGER,
	response_body TEXT
 );`

	_, err = DB.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Error creating table: %q: %s\n", err, sqlStmt)
	}
	return DB
}
