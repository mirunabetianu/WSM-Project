package main

import (
	"fmt"
	"os"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Please change this constants according to your setup
var (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "root"
	dbname   = "postgres"
)

var Database *gorm.DB

func connect() (error, *gorm.DB) {
	if(getEnv("POSTGRES_DB") != "") {
		dbname = getEnv("POSTGRES_DB")
	}
	
	if (getEnv("POSTGRES_USER") != "") {
		user = getEnv("POSTGRES_USER")
	} 
	
	if (getEnv("POSTGRES_PASSWORD") != "") {
		password = getEnv("POSTGRES_PASSWORD")
	}
	
	if (getEnv("POSTGRES_SERVICE_HOST") != "") {
		host = getEnv("POSTGRES_SERVICE_HOST")
	}

	var databaseInfo = fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	var err error
	Database, err = gorm.Open(postgres.Open(databaseInfo), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	Database.AutoMigrate(&User{})
	Database.AutoMigrate(&Payment{})

	return err, Database
}

func getEnv(key string) string{
	val, ok := os.LookupEnv(key)
	if !ok {
		return ""
	} else {
		return val
	}
}
