package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Please change this constants according to your setup

var dsn = "host=localhost user=postgres password=251219 dbname=postgres port=5433"

var Database *gorm.DB

func connect() (error, *gorm.DB) {
	var err error
	Database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	Database.AutoMigrate(&User{})
	Database.AutoMigrate(&Payment{})

	return err, Database
}
