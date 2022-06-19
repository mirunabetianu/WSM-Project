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

func GetEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return ""
	} else {
		return val
	}
}

func OpenPsqlConnection() *gorm.DB {

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
