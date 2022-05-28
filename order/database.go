package main

import (
	"database/sql"
	_ "database/sql"
	"fmt"
	_ "fmt"

	_ "github.com/lib/pq"
)

// Please change this constants according to your setup
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "root"
	dbname   = "postgres"
)

func openPsqlConnection() string {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return "connection open"
	}
	defer db.Close()
	return "connection failed"
}
