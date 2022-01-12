package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(gin.Recovery())

	router.GET("/status", status)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}

	log.Printf("Listening at %s", "http://localhost:8080")

	// Start serving traffic
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

// status endpoint to expose service livenessprobe
func status(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
