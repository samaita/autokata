package main

import (
	"fmt"
	"log"
	"time"

	DB "github.com/samaita/autokata/sql"
)

type Crawl struct {
	CrawlLogID int64
	DomainID   int64
	UpdateTime time.Time
}

func NewCrawl() Crawl {
	return Crawl{}
}

func (c *Crawl) getCrawlID() error {
	if c.CrawlLogID != 0 {
		return fmt.Errorf("Crawl ID is not 0, Got %d", c.CrawlLogID)
	}

	query := `
		INSERT INTO db_crawl_log (
			domain_id,
			update_time
		)
		VALUES ($1, $2)
	`

	tx, errQuery := DB.Collection.Main.Beginx()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	result, errQuery := tx.Exec(query, c.DomainID, time.Now().UTC().Format(time.RFC3339))
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	c.CrawlLogID, errQuery = result.LastInsertId()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	tx.Commit()

	return nil
}
