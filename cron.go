package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/robfig/cron"
)

func InitCronHourlyCrawlerByRSS() {
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

func InitCronHourlyCrawlerByURL() {
	c := cron.New()
	rule := "0 * * * *"
	log.Println("Cron HourlyCrawler By URL Scheduled:", rule)
	c.AddFunc(rule, func() {
		log.Println("Cron HourlyCrawler By URL Started")
		handleCronFetchBatchURL()
		log.Println("Cron HourlyCrawler By URL Completed")
	})
	c.Start()
}

func InitCronBotFetchUpdate() {
	c := cron.New()
	rule := "* * * * *"
	log.Println("Cron BotFetchUpdate Scheduled:", rule)
	c.AddFunc(rule, func() {
		log.Println("Cron BotFetchUpdate Started")
		handleCronBotFetchUpdate()
	})
	c.Start()
}

func handleCronBotFetchUpdate() {
	var (
		tu  TelegramUpdate
		err error
	)

	if tu, err = Bot.GetUpdate(); err != nil {
		log.Println(err)
		return
	}

	tu.GetLastUpdateID()
	lastInput := tu.GetLastUpdateMessage("/command")
	parsedLastInput := strings.ToLower(strings.ReplaceAll(lastInput, "/command ", ""))
	log.Println(lastInput, parsedLastInput)
	switch parsedLastInput {
	case "echo":
		Bot.SendMessage("Echo")
	case "list":
		Bot.SendMessage("/command echo: send echo. Dah itu aja")
	case "":
		log.Println("No New Input")
	default:
		log.Println(parsedLastInput, "Command Unknown, try another: /command list")
	}
	return
}

func handleCronFetchBatchRSS() {
	var (
		errGetAllDomain error
		totalNew        int
		titles          string
		listDomain      []Domain
	)

	if listDomain, errGetAllDomain = getDomainWithType(DOMAIN_FETCH_TYPE_RSS); errGetAllDomain != nil {
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
			isExist, errExist := feed.isURLExist()
			if errExist != nil {
				log.Println(errExist, feed.ArticleURL)
				continue
			}
			if isExist {
				continue
			}
			totalNew++
			titles = fmt.Sprintf(`%s

%s: %s`, titles, feed.ArticleTitle, feed.ArticleURL)
			feed.save()
		}

		if totalNew > 0 {
			Bot.SendMessage(fmt.Sprintf(`%d new article(s) from %s!%s`, totalNew, domain.DomainName, titles))
			titles = ""
			totalNew = 0
		}

	}
}

func handleCronFetchBatchURL() {
	var (
		errGetAllDomain, errGetCrawlID, errGetAllPos, errGetURLFeed error
		totalNew                                                    int
		titles                                                      string
		feeds                                                       []Feed
		listDomain                                                  []Domain
	)

	if listDomain, errGetAllDomain = getDomainWithType(DOMAIN_FETCH_TYPE_URL); errGetAllDomain != nil {
		log.Println(errGetAllDomain)
		return
	}

	for _, domain := range listDomain {
		crawl := NewCrawl()
		crawl.DomainID = domain.DomainID

		if errGetCrawlID = crawl.getCrawlID(); errGetCrawlID != nil {
			log.Println(errGetCrawlID, domain.DomainURL)
			continue
		}
		if errGetAllPos = domain.getAllPos(); errGetAllPos != nil {
			log.Println(errGetAllPos, domain.DomainURL)
			continue
		}
		if feeds, errGetURLFeed = domain.getURLFeedV2(); errGetURLFeed != nil {
			log.Println(errGetURLFeed, domain.DomainURL, "Retrying..")
			if feeds, errGetURLFeed = domain.getURLFeed(); errGetURLFeed != nil {
				log.Println(errGetURLFeed, domain.DomainURL, "Give Up..")
				continue
			}
		}

		for _, feed := range feeds {
			var (
				errExist error
				isExist  bool
			)

			feed.CrawlLogID = crawl.CrawlLogID
			feed.DomainID = crawl.DomainID

			if isExist, errExist = feed.isURLExist(); errExist != nil {
				log.Println(errExist, feed.ArticleURL)
				continue
			}
			if isExist {
				continue
			}

			totalNew++
			titles = fmt.Sprintf(`%s

%s: %s`, titles, feed.ArticleTitle, feed.ArticleURL)
			feed.save()
		}

		if totalNew > 0 {
			Bot.SendMessage(fmt.Sprintf(`%d new article(s) from %s!%s`, totalNew, domain.DomainName, titles))
			titles = ""
			totalNew = 0
		}
	}
}
