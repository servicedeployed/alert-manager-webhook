package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/alertme", handleAlert)

	r.GET("/healthz", func(c *gin.Context) {
		c.String(200, "Ok!")
	})

	r.Run()
}

func handleAlert(c *gin.Context) {
	// ensure JWT is present, fetch/refresh if needed

	// POST new alert payload to /alerting/webhook

}
