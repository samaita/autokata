package sql

import (
	"log"

	"github.com/samaita/autokata/sql/sqlx"
)

var Collection DBCollection

type DBCollection struct {
	Main sqlx.DB
}

func InitDB() {
	db, err := sqlx.New("sqlite3", "./db_autokata.db")
	if err != nil {
		log.Fatalln(err)
	}
	Collection.Main = db
	log.Println("[InitDB] DB Running....")
}
