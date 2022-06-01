package main

import (
	"fmt"
	_ "fmt"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Please change this constant according to your setup
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "root"
	dbname   = "postgres"
)

func openPsqlConnection() *gorm.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, errDb := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{})
	errMg := db.AutoMigrate(&Order{}, &Item{})
	if errDb != nil && errMg != nil {
		return nil
	}
	return db
}
