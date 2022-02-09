package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
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
