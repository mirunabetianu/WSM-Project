package utils

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
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

type User struct {
	gorm.Model
	Credit uint
}

type Payment struct {
	gorm.Model
	Status  byte
	OrderID uint
}

func OpenPsqlConnection() (error, *gorm.DB) {
	if GetEnv("POSTGRES_DB") != "" {
		dbname = GetEnv("POSTGRES_DB")
	}

	if GetEnv("POSTGRES_USER") != "" {
		user = GetEnv("POSTGRES_USER")
	}

	if GetEnv("POSTGRES_PASSWORD") != "" {
		password = GetEnv("POSTGRES_PASSWORD")
	}

	if GetEnv("POSTGRES_SERVICE_HOST") != "" {
		host = GetEnv("POSTGRES_SERVICE_HOST")
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

func GetEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return ""
	} else {
		return val
	}
}
