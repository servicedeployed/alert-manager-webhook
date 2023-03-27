package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/prometheus/alertmanager/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var AuthToken string
var WebhookUrl string

const appName = "alertmanager-webhook"

func main() {
	// Default to false
	debug := false
	// Grab the ENV
	debugENV := os.Getenv("DEBUG")
	// Check if ENV was passed in, default to false
	debugBool, err := strconv.ParseBool(debugENV)
	if err == nil {
		debug = debugBool
	}

	// Set up gin routing
	r := gin.Default()

	// Check if debug mode is on, if so, skip kubernetes logic
	if debug {
		fmt.Println("debug: debug mode is on, skipping kubernetes initialization")
	} else {
		fmt.Println("kubernetes: initializing local kubernetes config")
		initializeKubernetesConfig()
		fmt.Println("kubernetes: kubeconfig inititialized")
		r.POST("/alertme", handleAlert)
		fmt.Println("gin: /alertme alert-webhook route registered")
	}

	r.GET("/healthz", func(c *gin.Context) {
		c.String(200, "Ok!")
	})
	fmt.Println("gin: /healthz healthcheck route registered")
	fmt.Printf("%s: now listening for events", appName)
	r.Run()
}

func initializeKubernetesConfig() {
	// Load local config for pod
	localconfig, err := rest.InClusterConfig()
	if err != nil {
		fmt.Printf("kubernetes-error: could not get in-cluster config %v", err)
	}

	localclientset, err := kubernetes.NewForConfig(localconfig)
	if err != nil {
		fmt.Printf("kubernetes-error: could not get local clientset %v", err)
	}
	fmt.Println("kubernetes: local config loaded")

	// Load credentials for namespace
	namespace := "default"
	apiTokenNamespace := os.Getenv("API_TOKEN_SECRET_NAMESPACE")

	if apiTokenNamespace != "default" {
		namespaceBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
		if err != nil {
			fmt.Printf("kubernetes-error: could not get deployed namespace %v", err)
		} else {
			namespace = string(namespaceBytes)
		}
	} else {
		namespace = apiTokenNamespace
	}

	secretName := os.Getenv("API_TOKEN_SECRET_NAME")
	secret, err := localclientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("kubernetes-error: could not get secret %s %v", secretName, err)
	}
	fmt.Printf("kubernetes: found secret: %s/%s\n", namespace, secretName)

	authToken, ok := secret.Data["key"]
	if !ok {
		fmt.Printf("kubernetes-error: secret %s does not contain key 'value'", secretName)
		return
	}
	AuthToken = string(authToken)
	fmt.Printf("kubernetes: %s auth token loaded\n", appName)

	webhookUrl, ok := secret.Data["url"]
	if !ok {
		// couldn't find the value inside the secret
		fmt.Printf("kubernetes-error: secret %s does not contain key 'url'", secretName)
		return
	}
	WebhookUrl = string(webhookUrl)
	fmt.Printf("kubernetes: %s webhook url loaded\n", appName)
	fmt.Printf("%s: using webhook, %s\n", appName, webhookUrl)
}

func handleAlert(c *gin.Context) {
	// process request data
	data := &template.Data{}
	if err := json.NewDecoder(c.Request.Body).Decode(&data); err != nil {
		fmt.Printf("gin: bad input %v", err)
		c.String(400, "Bad Input")
		return
	}

	// POST new alert payload to /alerting/webhook

	result := map[string]interface{}{}
	client := resty.New()
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetAuthToken(AuthToken).
		SetBody(map[string]interface{}{"alerts": data.Alerts.Firing()}).
		SetResult(result).
		Post(WebhookUrl)
	if err != nil {
		fmt.Printf("gin: failed to send alerts to backend API %v", err)
	}

	if resp.StatusCode() != 200 {
		fmt.Printf("gin: unsuccesful response from API\n%+v\nBody: %s", resp.RawResponse, string(resp.Body()))
	}

	c.String(200, "Ok")
}
