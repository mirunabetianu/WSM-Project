package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Please change this constants according to your setup
const (
	host     = "localhost"
	port     = 5433
	user     = "postgres"
	password = "root"
	dbname   = "postgres"
)

var Database *gorm.DB

func connect() (error, *gorm.DB) {
	var err error
	Database, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		panic("failed to connect to database")
	}

	Database.AutoMigrate(&User{})
	Database.AutoMigrate(&Payment{})

	return err, Database
}
