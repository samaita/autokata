package main

import (
	"time"

	"github.com/gin-gonic/gin"
)

func basicMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(keyTimeStart, time.Now())
	}
}
