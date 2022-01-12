package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/mmcdole/gofeed"
	DB "github.com/samaita/autokata/sql"
)

const (
	ArticleExist  = 1
	ArticleStored = 2
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
		f.Status = ArticleExist
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

func NewFeed() Feed {
	return Feed{}
}

func (f *Feed) isURLExist() (bool, error) {
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
			article_cover_image,
			article_summary,
			article_content,
			article_publish_time,
			create_time,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	tx, errQuery := DB.Collection.Main.Beginx()
	if errQuery != nil {
		log.Println(errQuery, query)
		return errQuery
	}
	result, err := tx.Exec(query, f.CrawlLogID, f.DomainID, f.ArticleTitle, f.ArticleURL, f.ArticleCoverImage, f.ArticleSummary, f.ArticleContent, f.ArticlePublishTime.Format(time.RFC3339), time.Now().UTC().Format(time.RFC3339), f.Status)
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

func (f *Feed) Load() error {
	var (
		err error
	)

	if err = f.LoadFromDB(); err != nil {
		return err
	}

	if f.Status < ArticleStored {
		if err = f.Populate(); err != nil {
			return err
		}
	}

	return nil
}

func (f *Feed) Populate() error {
	var (
		s string
	)

	c := colly.NewCollector()

	// Find and visit all links
	c.OnHTML(".ArticlePage-mainContent", func(e *colly.HTMLElement) {
		a := strings.Replace(e.Text, "  ", "", -1)
		a = strings.Replace(a, "\n", "", -1)
		s = fmt.Sprintf("%s%s", s, a)
		f.ArticleContent = s
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	log.Println("DO", f.ArticleURL)

	c.Visit(f.ArticleURL)
	return nil
}

func (f *Feed) LoadFromDB() error {
	var (
		errQuery, errParse error
		tp, tc             string
	)

	query := "SELECT domain_id, article_id, article_title, article_url, article_cover_image, article_summary, article_content, article_publish_time, create_time, status FROM db_article WHERE article_id = $1"
	errQuery = DB.Collection.Main.QueryRowx(query, f.ArticleID).Scan(&f.DomainID, &f.ArticleID, &f.ArticleTitle, &f.ArticleURL, &f.ArticleCoverImage, &f.ArticleSummary, &f.ArticleContent, &tp, &tc, &f.Status)
	if errQuery != nil && errQuery != sql.ErrNoRows {
		log.Println(errQuery, query)
		return errQuery
	}
	if f.ArticlePublishTime, errParse = time.Parse(time.RFC3339, tp); errParse != nil {
		log.Println(errParse, tp)
		return errParse
	}
	if f.CreateTime, errParse = time.Parse(time.RFC3339, tc); errParse != nil {
		log.Println(errParse, tc)
		return errParse
	}

	return nil
}
