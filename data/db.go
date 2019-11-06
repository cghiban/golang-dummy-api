package data

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

/*
connStr := "user=pqgotest dbname=pqgotest sslmode=verify-full"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
*/

// InitDB(<filepath>)
func InitDB(connectionString string) *sql.DB {
	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	if db == nil {
		log.Fatal("unable to get a db connection")
	}
	return db
}
