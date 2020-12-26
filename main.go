package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	DB "github.com/samaita/autokata/sql"
)

var (
	APITimeout = 60 * time.Second

	Bot              TelegramBot
	botToken, chatID string
	err              error
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	DB.InitDB()

	// set to multiple value, separate init to another file
	if botToken, err = getKV("telegram_bot_token"); err != nil {
		log.Fatalln(err)
	}
	if chatID, err = getKV("telegram_chat_id"); err != nil {
		log.Fatalln(err)
	}

	Bot = NewTelegramBot(botToken, chatID)
}

func main() {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	defaultRoute := r.Group("/")
	defaultRoute.Use(
		basicMiddleware(),
	)
	{
		defaultRoute.POST("/domain/add", handleDomainAdd)
		defaultRoute.POST("/domain/remove", handleDomainRemove)
		defaultRoute.GET("/domain/list", handleDomainList)
		defaultRoute.GET("/feed/list", handleFeedList)
		defaultRoute.GET("/feed/fetch", handleFeedFetch)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		InitCronHourlyCrawlerByRSS()
		InitCronHourlyCrawlerByURL()
		// InitCronBotFetchUpdate()
		<-c
		log.Println("APP STOPPED")
		os.Exit(1)
	}()

	r.Run(":3000")
}
