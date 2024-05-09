package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/omc/bonsai-api-go/v2/bonsai"
)

func main() {
	apiKey := os.Getenv("BONSAI_API_KEY")
	apiToken := os.Getenv("BONSAI_API_TOKEN")

	client := bonsai.NewClient(
		bonsai.WithCredentialPair(
			bonsai.CredentialPair{
				AccessKey:   bonsai.AccessKey(apiKey),
				AccessToken: bonsai.AccessToken(apiToken),
			},
		),
	)

	clusters, err := client.Cluster.All(context.Background())
	if err != nil {
		log.Fatalf("error listing clusters: %s\n", err)
	}

	asJson, err := json.MarshalIndent(clusters, "", "    ")
	log.Printf("Found %d clusters! Details: %v\n", len(clusters), string(asJson))
}
