package utils

import (
	_ "database/sql"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "Miruna999"
	dbname   = "wsm"
)

type Item struct {
	gorm.Model
	Stock uint `gorm:"default:0"`
	Price uint `gorm:"default:0"`
}

func getEnv(key string) string{
	val, ok := os.LookupEnv(key)
	if !ok {
		return ""
	} else {
		return val
	}
}

func OpenPsqlConnection() *gorm.DB {

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
	
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, errDb := gorm.Open(postgres.Open(psqlInfo), &gorm.Config{})
	errMg := db.AutoMigrate(&Item{})
	if errDb != nil && errMg != nil {
		return nil
	}
	return db
}