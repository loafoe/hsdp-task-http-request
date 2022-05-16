package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/philips-software/go-hsdp-api/iam"
	"github.com/spf13/viper"
)

func main() {
	viper.SetEnvPrefix("request")
	viper.SetDefault("method", "POST")
	viper.SetDefault("body", "")
	viper.SetDefault("url", "")
	viper.SetDefault("username", "")
	viper.SetDefault("password", "")
	viper.AutomaticEnv()

	client := resty.New()

	r := client.R()

	username := viper.GetString("username")
	password := viper.GetString("password")
	if username != "" {
		r = r.SetBasicAuth(username, password)
	}
	// Get headers
	for _, entry := range os.Environ() {
		parts := strings.Split(entry, "=")
		if len(parts) < 2 || !strings.HasPrefix(parts[0], "REQUEST_HEADER_") {
			continue
		}
		headerName := strings.Replace(parts[0], "REQUEST_HEADER_", "", 1)
		canonicalHeader := strings.Replace(headerName, "_", "-", -1)
		headerValue := parts[1]
		r = r.SetHeader(canonicalHeader, headerValue)
	}
	r = detectAndUseServiceIdentity(r)

	body := viper.GetString("body")
	if body != "" {
		r = r.SetBody(body)
	}
	method := viper.GetString("method")
	url := viper.GetString("url")

	fmt.Printf("%v %v\n", method, url)
	resp, err := r.Execute(method, url)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	fmt.Printf("response: [%v]\n", resp)
}

func detectAndUseServiceIdentity(req *resty.Request) *resty.Request {
	serviceID := os.Getenv("HSDP_IAM_SERVICE_ID")
	privateKey := os.Getenv("HSDP_IAM_SERVICE_PRIVATE_KEY")
	region := os.Getenv("HSDP_REGION")
	environment := os.Getenv("HSDP_ENVIRONMENT")
	if serviceID == "" || privateKey == "" || region == "" || environment == "" {
		return req
	}
	fmt.Printf("iam: found service credentials, using them to generate accessToken\n")
	client, err := iam.NewClient(nil, &iam.Config{
		Region:      region,
		Environment: environment,
	})
	if err != nil {
		fmt.Printf("iam: error creating client: %v\n", err)
		return req
	}
	err = client.ServiceLogin(iam.Service{
		ServiceID:  serviceID,
		PrivateKey: privateKey,
	})
	if err != nil {
		fmt.Printf("iam: error logging in: %v\n", err)
		return req
	}
	token, err := client.Token()
	if err != nil {
		fmt.Printf("iam: error getting token: %v\n", err)
		return req
	}
	req.SetHeader("Authorization", "Bearer "+token)
	return req
}
