package main

import (
	"context"
	"log"
	"time"

	DB "github.com/samaita/autokata/sql"
)

var (
	APITimeout = 60 * time.Second
)

// type Item struct {
// 	Title         string                   `json:"title,omitempty"`
// 	Link          string                   `json:"link,omitempty"`
// 	Description   string                   `json:"description,omitempty"`
// 	Content       string                   `json:"content,omitempty"`
// 	Author        string                   `json:"author,omitempty"`
// 	Categories    []*Category              `json:"categories,omitempty"`
// 	Comments      string                   `json:"comments,omitempty"`
// 	Enclosure     *Enclosure               `json:"enclosure,omitempty"`
// 	GUID          *GUID                    `json:"guid,omitempty"`
// 	PubDate       string                   `json:"pubDate,omitempty"`
// 	PubDateParsed *time.Time               `json:"pubDateParsed,omitempty"`
// 	Source        *Source                  `json:"source,omitempty"`
// }

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	DB.InitDB()
}

func main() {
	var (
		errGetAllDomain error
		totalNew        int
	)

	listDomain, errGetAllDomain := GetAllDomain()
	if errGetAllDomain != nil {
		log.Println(errGetAllDomain)
		return
	}

	for _, domain := range listDomain {
		crawl := NewCrawl()
		crawl.DomainID = domain.DomainID

		errGetCrawlID := crawl.GetCrawlID()
		if errGetCrawlID != nil {
			log.Println(errGetCrawlID, domain.DomainURL)
			continue
		}

		feeds, errGetRSSFeed := GetRSSFeed(context.Background(), domain.FeedsURL)
		if errGetRSSFeed != nil {
			log.Println(errGetRSSFeed, domain.DomainURL)
			continue
		}

		for _, feed := range feeds {
			feed.CrawlLogID = crawl.CrawlLogID
			feed.DomainID = crawl.DomainID
			isExist, errExist := feed.IsExist()
			if errExist != nil {
				log.Println(errExist, feed.ArticleURL)
				continue
			}
			if isExist {
				continue
			}
			totalNew++
			feed.Save()
		}

		if totalNew > 0 {
			log.Println(totalNew, "new article(s)!")
		}
	}

}
