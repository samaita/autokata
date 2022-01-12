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

func setKV(k, v string) error {
	var (
		errQuery error
	)

	query := `UPDATE kv SET value = $1 WHERE key = $2`
	tx, errQuery := DB.Collection.Main.Beginx()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	if _, errQuery = tx.Exec(query, v, k); errQuery != nil {
		return errQuery
	}
	tx.Commit()
	return nil
}
