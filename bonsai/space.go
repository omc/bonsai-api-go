package bonsai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	spaceAPIBasePath = "/spaces"
)

// CloudProvider contains details about the cloud provider and region
// attributes.
type CloudProvider struct {
	Provider string `json:"provider"`
	Region   string `json:"region"`
}

// Space represents the server groups and geographic regions available to their
// account, where clusters may be provisioned.
type Space struct {
	Path           string        `json:"path"`
	PrivateNetwork bool          `json:"private_network"`
	Cloud          CloudProvider `json:"cloud"`
}

// SpacesResultList is a wrapper around a slice of
// Spaces for json unmarshaling.
type SpacesResultList struct {
	Spaces []Space `json:"spaces"`
}

// SpaceClient is a client for the Spaces API.
type SpaceClient struct {
	*Client
}

// list returns a list of Spaces for the page specified,
// by performing a GET request against [spaceAPIBasePath].
//
// Note: Pagination is not currently supported.
func (c *SpaceClient) list(ctx context.Context) ([]Space, *Response, error) {
	var (
		req  *http.Request
		resp *Response
		err  error
	)
	// Let's make some initial capacity to reduce allocations
	results := SpacesResultList{
		Spaces: make([]Space, 0, defaultResponseCapacity),
	}

	req, err = c.NewRequest(ctx, "GET", spaceAPIBasePath, nil)
	if err != nil {
		return nil, nil, err
	}

	resp, err = c.Do(ctx, req)
	if err != nil {
		return nil, resp, fmt.Errorf("client.do failed: %w", err)
	}

	if err = json.Unmarshal(resp.BodyBuf.Bytes(), &results); err != nil {
		return nil, resp, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	return results.Spaces, resp, nil
}

// All lists all Spaces from the Spaces API.
func (c *SpaceClient) All(ctx context.Context) ([]Space, error) {
	// No pagination support as yet
	results, _, err := c.list(ctx)
	if err != nil {
		return results, fmt.Errorf("client.list failed: %w", err)
	}
	return results, err
}
