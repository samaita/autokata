package main

import (
	"database/sql"
	"log"

	DB "github.com/samaita/autokata/sql"
)

func getKV(k string) (string, error) {
	var (
		v        string
		errQuery error
	)

	query := `SELECT value FROM kv WHERE key = $1`
	errQuery = DB.Collection.Main.QueryRowx(query, k).Scan(&v)
	if errQuery != nil && errQuery != sql.ErrNoRows {
		log.Println(errQuery, query)
		return v, errQuery
	}
	return v, nil
}
