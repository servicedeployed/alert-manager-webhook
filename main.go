package main

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/alertmanager/template"
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
	// process request data
	data := &template.Data{}
	if err := json.NewDecoder(c.Request.Body).Decode(&data); err != nil {
		c.String(400, "Bad Input")
		return
	}

	// ensure JWT is present, fetch/refresh if needed

	// POST new alert payload to /alerting/webhook
	fmt.Printf("%+v\n", data)

	c.String(200, "Ok")
}
