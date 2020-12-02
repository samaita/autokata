package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	DB "github.com/samaita/autokata/sql"
)

var (
	APITimeout = 60 * time.Second
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	DB.InitDB()
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
	}

	go func() {
		InitCronHourlyCrawler()
	}()

	r.Run(":3000")
}
