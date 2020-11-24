package sqlx

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type db struct {
	Conn *sqlx.DB
}

// New SQLX connection
func New(driverName, dataSourceName string) (DB, error) {
	result, err := sqlx.Connect(driverName, dataSourceName)
	if err != nil {
		log.Printf("[SQLX][Connect] %v", err)
		return nil, err
	}

	connection := &db{
		Conn: result,
	}

	return DB(connection), nil
}
