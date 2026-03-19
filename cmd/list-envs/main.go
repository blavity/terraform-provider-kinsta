package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/blavity/terraform-provider-kinsta/internal/client"
)

func main() {
	if len(os.Args) < 4 {
		log.Fatal("Usage: list-envs <api_key> <company_id> <site_id>")
	}

	apiKey := os.Args[1]
	companyID := os.Args[2]
	siteID := os.Args[3]

	c := client.New(apiKey, companyID)

	// Get the site to see its environments
	site, err := c.GetWordPressSite(context.Background(), siteID)
	if err != nil {
		log.Fatalf("Failed to get site: %v", err)
	}

	// Pretty print the site response
	data, _ := json.MarshalIndent(site, "", "  ")
	fmt.Println(string(data))
}
