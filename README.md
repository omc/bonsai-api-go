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
go get github.com/omc/bonsai-api-go/v1
```

## Example

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/omc/bonsai-api-go/v1/bonsai"
)

func main() {
	token, err := bonsai.NewToken("TestToken")
	if err != nil {
		log.Fatal(fmt.Errorf("invalid token: %w", err))
	}
	
	client := bonsai.NewClient(
		bonsai.WithToken(token),
	)

	clusters, _, err := client.Clusters.All(context.Background())
	if err != nil {
		log.Fatalf("error listing clusters: %s\n", err)
	}
	log.Printf("Found %d clusters!\n", len(clusters))
}
```

## Contributing

### Pre-commit

This project uses [pre-commit](https://pre-commit.com/) to lint and store 3rd-party dependency licenses.
Installation instructions are available on the [pre-commit](https://pre-commit.com/) website!

To verify your installation, run this project's pre-commit hooks against all files:

```shell
pre-commit run --all-files
```

#### Go-licenses pre-commit hook

Windows users: Ensure that you have `C:\Program Files\Git\usr\bin` added
to your `PATH`!