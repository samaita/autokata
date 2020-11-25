package main

import (
	"log"
	"time"

	DB "github.com/samaita/autokata/sql"
)

type Domain struct {
	DomainID   int64
	DomainName string
	DomainURL  string
	FeedsURL   string
	CreateTime time.Time
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
		errScan = rows.Scan(&d.DomainID, &d.DomainName, &d.DomainURL, &d.FeedsURL, &t)
		if errScan != nil {
			log.Println(errScan)
			continue
		}
		d.CreateTime, errParse = time.Parse(time.RFC3339, t)
		if errParse != nil {
			log.Println(errParse, t)
			continue
		}
		Domains = append(Domains, d)
	}

	return Domains, nil
}
