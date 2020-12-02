package main

import (
	"context"
	"log"

	"github.com/robfig/cron"
)

func InitCronHourlyCrawler() {
	c := cron.New()
	rule := "0 * * * *"
	log.Println("Cron HourlyCrawler Scheduled:", rule)
	c.AddFunc(rule, func() {
		log.Println("Cron HourlyCrawler Started")
		handleCronFetchBatchRSS()
		log.Println("Cron HourlyCrawler Completed")
	})
	c.Start()
}

func handleCronFetchBatchRSS() {
	var (
		errGetAllDomain error
		totalNew        int
	)

	listDomain, errGetAllDomain := getAllDomain()
	if errGetAllDomain != nil {
		log.Println(errGetAllDomain)
		return
	}

	for _, domain := range listDomain {
		crawl := NewCrawl()
		crawl.DomainID = domain.DomainID

		errGetCrawlID := crawl.getCrawlID()
		if errGetCrawlID != nil {
			log.Println(errGetCrawlID, domain.DomainURL)
			continue
		}

		feeds, errGetRSSFeed := getRSSFeed(context.Background(), domain.FeedsURL)
		if errGetRSSFeed != nil {
			log.Println(errGetRSSFeed, domain.DomainURL)
			continue
		}

		for _, feed := range feeds {
			feed.CrawlLogID = crawl.CrawlLogID
			feed.DomainID = crawl.DomainID
			isExist, errExist := feed.isExist()
			if errExist != nil {
				log.Println(errExist, feed.ArticleURL)
				continue
			}
			if isExist {
				continue
			}
			totalNew++
			feed.save()
		}

		if totalNew > 0 {
			log.Println(totalNew, "new article(s)! from ", domain.DomainName)
		}
	}
}
