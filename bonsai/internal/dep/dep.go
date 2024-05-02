// Package dep imports unused, but licensed, dependencies
// for capture by tooling like "go mod vendor" and
// github.com/google/go-licenses
package dep

import (
	// bonsai.Client is based on hcloud's implementation.
	_ "github.com/hetznercloud/hcloud-go/v2/hcloud"
)
