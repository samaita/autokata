package main

import (
	"database/sql"
	"log"
	"time"

	DB "github.com/samaita/autokata/sql"
)

type Domain struct {
	DomainID   int64     `json:"domain_id"`
	DomainName string    `json:"domain_name"`
	DomainURL  string    `json:"domain_url"`
	FeedsURL   string    `json:"feeds_url"`
	CreateTime time.Time `json:"create_time"`
}

func getAllDomain() ([]Domain, error) {
	var (
		errParse, errQuery, errScan error
		Domains                     []Domain
	)

	query := "SELECT domain_id, domain_name, domain_url, feeds_url, create_time FROM db_domain"
	rows, errQuery := DB.Collection.Main.Queryx(query)
	if errQuery != nil {
		log.Println(errQuery, query)
		return Domains, errQuery
	}
	defer rows.Close()

	for rows.Next() {
		var (
			d Domain
			t string
		)
		if errScan = rows.Scan(&d.DomainID, &d.DomainName, &d.DomainURL, &d.FeedsURL, &t); errScan != nil {
			log.Println(errScan)
			continue
		}
		if d.CreateTime, errParse = time.Parse(time.RFC3339, t); errParse != nil {
			log.Println(errParse, t)
			continue
		}
		Domains = append(Domains, d)
	}

	return Domains, nil
}

func NewDomain() Domain {
	return Domain{}
}

func (d *Domain) validate() bool {
	if d.DomainName == "" || d.DomainURL == "" || d.FeedsURL == "" {
		return false
	}
	// fix url from https://gundam.org/ to https://gundam.org
	return true
}

func (d *Domain) add() error {
	query := `
		INSERT INTO db_domain (
			domain_name,
			domain_url,
			feeds_url,
			create_time
		)
		VALUES ($1, $2, $3, $4)
	`
	tx, errQuery := DB.Collection.Main.Beginx()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	result, err := tx.Exec(query, d.DomainName, d.DomainURL, d.FeedsURL, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		return err
	}
	if d.DomainID, err = result.LastInsertId(); errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	tx.Commit()
	return nil
}

func (d *Domain) isExist() (bool, error) {
	var (
		b        bool
		s        int
		errQuery error
	)

	query := `SELECT 1 FROM db_domain WHERE domain_url = $1`
	errQuery = DB.Collection.Main.QueryRowx(query, d.DomainURL).Scan(&s)
	if errQuery != nil && errQuery != sql.ErrNoRows {
		log.Println(errQuery, query)
		return b, errQuery
	}
	b = s == 1
	return b, nil
}

func (d *Domain) remove() error {
	query := `DELETE FROM db_domain WHERE domain_url = $1`
	tx, errQuery := DB.Collection.Main.Beginx()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	_, err := tx.Exec(query, d.DomainURL)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
