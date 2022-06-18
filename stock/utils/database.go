package utils

import (
	_ "database/sql"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
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

func OpenPsqlConnection() *gorm.DB {
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
