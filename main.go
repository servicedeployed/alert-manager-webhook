package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/prometheus/alertmanager/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var AuthToken string
var WebhookUrl string

func main() {

	localconfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("could not get in-cluster config %v", err)

	}

	localclientset, err := kubernetes.NewForConfig(localconfig)
	if err != nil {
		fmt.Printf("could not get local clientset %v", err)

	}

	secret, err := localclientset.CoreV1().Secrets("default").Get(context.TODO(), "cluster-token", metav1.GetOptions{})
	if err != nil {
		fmt.Printf("could not get secret cluster-token %v", err)

	}

	authToken, ok := secret.Data["key"]
	if !ok {
		// couldn't find the value inside the secret
		fmt.Println("could not find a key secret value, exiting")
		return
	}
	AuthToken = string(authToken)

	webhookUrl, ok := secret.Data["url"]
	if !ok {
		// couldn't find the value inside the secret
		fmt.Println("could not find a url secret value, exiting")
		return
	}
	WebhookUrl = string(webhookUrl)

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

	// POST new alert payload to /alerting/webhook

	result := map[string]interface{}{}
	client := resty.New()
	_, err := client.R().
		SetHeader("Accept", "application/json").
		SetAuthToken(AuthToken).
		SetBody(map[string]interface{}{"alerts": data.Alerts.Firing()}).
		SetResult(result).
		Post(WebhookUrl)
	if err != nil {
		fmt.Printf("failed to send alerts to backend API %v", err)
	}

	c.String(200, "Ok")
}
