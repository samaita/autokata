package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/mmcdole/gofeed"
	DB "github.com/samaita/autokata/sql"
)

const (
	ArticlePublished = 1
)

type Feed struct {
	ArticleID          int64     `json:"article_id,omitempty"`
	CrawlLogID         int64     `json:"-"`
	DomainID           int64     `json:"-"`
	ArticleTitle       string    `json:"article_title,omitempty"`
	ArticleURL         string    `json:"article_url,omitempty"`
	ArticleCoverImage  string    `json:"article_cover_image,omitempty"`
	ArticleSummary     string    `json:"article_summary,omitempty"`
	ArticleContent     string    `json:"article_content,omitempty"`
	ArticleListImage   string    `json:"-"`
	ArticlePublishTime time.Time `json:"article_publish_time,aomitempty"`
	CreateTime         time.Time `json:"-"`
	UpdateTime         time.Time `json:"-"`
	Status             int64     `json:"-"`
}

// GetRSSFeed obtain feed then translate to our definition
func getRSSFeed(ctx context.Context, RSSUrl string) ([]Feed, error) {
	var (
		Feeds []Feed
	)
	ctx, cancel := context.WithTimeout(context.Background(), APITimeout)
	defer cancel()

	fp := gofeed.NewParser()
	feed, errParse := fp.ParseURLWithContext(RSSUrl, ctx)
	if errParse != nil {
		log.Println(errParse, RSSUrl)
		return Feeds, errParse
	}

	for _, item := range feed.Items {
		var f Feed
		f.ArticleURL = item.Link
		f.ArticleTitle = item.Title
		f.ArticleSummary = removeHtmlTag(item.Description)
		f.ArticlePublishTime = *item.PublishedParsed
		f.CreateTime = time.Now().UTC()
		f.Status = ArticlePublished
		Feeds = append(Feeds, f)
	}

	return Feeds, nil
}

func getAllFeed() ([]Feed, error) {
	var (
		errParse, errQuery, errScan error
		Feeds                       []Feed
	)

	query := "SELECT domain_id, article_id, article_title, article_url, article_publish_time, create_time FROM db_article"
	rows, errQuery := DB.Collection.Main.Queryx(query)
	if errQuery != nil {
		log.Println(errQuery, query)
		return Feeds, errQuery
	}
	defer rows.Close()

	for rows.Next() {
		var (
			f      Feed
			tp, tc string
		)
		if errScan = rows.Scan(&f.DomainID, &f.ArticleID, &f.ArticleTitle, &f.ArticleURL, &tp, &tc); errScan != nil {
			log.Println(errScan)
			continue
		}
		if f.ArticlePublishTime, errParse = time.Parse(time.RFC3339, tp); errParse != nil {
			log.Println(errParse, tp)
			continue
		}
		if f.CreateTime, errParse = time.Parse(time.RFC3339, tc); errParse != nil {
			log.Println(errParse, tc)
			continue
		}
		Feeds = append(Feeds, f)
	}

	return Feeds, nil

}

func (f *Feed) isExist() (bool, error) {
	var (
		b        bool
		s        int
		errQuery error
	)

	query := "SELECT 1 FROM db_article WHERE article_url = $1"
	errQuery = DB.Collection.Main.QueryRowx(query, f.ArticleURL).Scan(&s)
	if errQuery != nil && errQuery != sql.ErrNoRows {
		log.Println(errQuery, query)
		return b, errQuery
	}
	b = s == 1
	return b, nil
}

func (f *Feed) save() error {
	query := `
		INSERT INTO db_article (
			crawl_log_id,
			domain_id,
			article_title,
			article_url,
			article_summary,
			article_publish_time,
			create_time,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	tx, errQuery := DB.Collection.Main.Beginx()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	result, err := tx.Exec(query, f.CrawlLogID, f.DomainID, f.ArticleTitle, f.ArticleURL, f.ArticleSummary, f.ArticlePublishTime.Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339), f.Status)
	if err != nil {
		return err
	}
	f.ArticleID, err = result.LastInsertId()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	tx.Commit()

	return nil
}
