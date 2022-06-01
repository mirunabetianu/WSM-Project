package database

import (
	"fmt"
	_ "fmt"
	"github.com/lib/pq"
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

type Order struct {
	gorm.Model
	Paid      bool          `gorm:"type:bool;default:false"`
	UserId    string        `gorm:"type:varchar;not null"`
	TotalCost int           `gorm:"type:bigint;default:0"`
	Items     pq.Int64Array `gorm:"type:integer[]"`
}

func OpenPsqlConnection() *gorm.DB {
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
