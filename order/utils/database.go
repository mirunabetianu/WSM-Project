package utils

import (
	"fmt"
	_ "fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

// Please change this constant according to your setup
var (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "root"
	dbname   = "postgres"
)

type Order struct {
	gorm.Model
	Paid      bool          `gorm:"type:bool;default:false"`
	UserId    string        `gorm:"type:varchar;not null"`
	TotalCost int           `gorm:"type:bigint;default:0"`
	Items     pq.Int64Array `gorm:"type:integer[]"`
}

type Item struct {
	gorm.Model
	Stock uint `json:"stock"`
	Price uint `json:"price"`
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
	errMg := db.AutoMigrate(&Order{})
	if errDb != nil && errMg != nil {
		return nil
	}
	return db
}

func GetEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return ""
	} else {
		return val
	}
}
