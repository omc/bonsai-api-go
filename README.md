# bonsai-api-go: Bonsai Cloud Go API Client

![Bonsai | Fully Managed Elasticsearch & OpenSearch](doc/assets/bonsai.png)

This is the Go API Client for [Bonsai Cloud](https://bonsai.io/) - The only 
managed Elasticsearch, OpenSearch, and SolrCloud platform that provides the 
support of a search engineering team, but at a fraction of the cost!

## Documentation

Full documentation is available on godoc at https://pkg.go.dev/github.com/omc/bonsai-api-go/v1/bonsai/

### Self-hosting the docs

1. Install [godoc](https://pkg.go.dev/golang.org/x/tools/cmd/godoc)
   ```shell
   go install golang.org/x/tools/cmd/godoc@latest
   ```
2. Run `godoc` from the project's root directory:
   ```shell
   godoc -http:6060
   ```

## Installation

```shell
go get github.com/omc/bonsai-api-go/bonsai
```

## Example

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/omc/bonsai-api-go/bonsai"
)

func main() {
	// Fetch API Key and Token from environment variables
	apiKey := os.Getenv("BONSAI_API_KEY")
	apiToken := os.Getenv("BONSAI_API_TOKEN")

	// Create a new Bonsai API Client
	client := bonsai.NewClient(
		bonsai.WithCredentialPair(
			bonsai.CredentialPair{
				AccessKey:   bonsai.AccessKey(apiKey),
				AccessToken: bonsai.AccessToken(apiToken),
			},
		),
	)

	// Fetch all of our clusters
	clusters, err := client.Cluster.All(context.Background())
	if err != nil {
		log.Fatalf("error listing clusters: %s\n", err)
	}
	log.Printf("Found %d clusters! Details: %v\n", len(clusters), clusters)
}
```
