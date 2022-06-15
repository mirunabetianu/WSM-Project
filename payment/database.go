package main

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Please change this constants according to your setup
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "root"
	dbname   = "postgres"
)

var databaseInfo = fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
var Database *gorm.DB

func connect() (error, *gorm.DB) {
	var err error
	Database, err = gorm.Open(postgres.Open(databaseInfo), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	Database.AutoMigrate(&User{})
	Database.AutoMigrate(&Payment{})

	return err, Database
}
