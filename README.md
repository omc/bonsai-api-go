# bonsai-api-go: Bonsai Cloud Go API Client

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