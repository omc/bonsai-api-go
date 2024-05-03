package bonsai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"
)

const (
	SpaceAPIBasePath = "/spaces"
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
	Cloud          CloudProvider `json:"cloud,omitempty"`

	// The geographic region in which the cluster is running.
	Region string `json:"region,omitempty"`
	// A URI to retrieve more information about this Release.
	URI string `json:"uri,omitempty"`
}

// SpacesResultList is a wrapper around a slice of
// Spaces for json unmarshaling.
type SpacesResultList struct {
	Spaces []Space `json:"spaces,omitempty"`
}

// SpaceClient is a client for the Spaces API.
type SpaceClient struct {
	*Client
}

type SpaceListOptions struct {
	listOpts
}

func (o SpaceListOptions) values() url.Values {
	return o.listOpts.values()
}

// list returns a list of Spaces for the page specified,
// by performing a GET request against [spaceAPIBasePath].
//
// Note: Pagination is not currently supported.
func (c *SpaceClient) list(ctx context.Context, opt SpaceListOptions) ([]Space, *Response, error) {
	var (
		req    *http.Request
		reqURL *url.URL
		resp   *Response
		err    error
	)
	// Let's make some initial capacity to reduce allocations
	results := SpacesResultList{
		Spaces: make([]Space, 0, defaultResponseCapacity),
	}

	reqURL, err = url.Parse(SpaceAPIBasePath)
	if err != nil {
		return results.Spaces, nil, fmt.Errorf("cannot parse relative url from basepath (%s): %w", SpaceAPIBasePath, err)
	}

	// Conditionally set options if we received any
	if !reflect.ValueOf(opt).IsZero() {
		reqURL.RawQuery = opt.values().Encode()
	}

	req, err = c.NewRequest(ctx, "GET", reqURL.String(), nil)
	if err != nil {
		return results.Spaces, nil, fmt.Errorf("creating new http request for URL (%s): %w", reqURL.String(), err)
	}

	resp, err = c.Do(ctx, req)
	if err != nil {
		return results.Spaces, resp, fmt.Errorf("client.do failed: %w", err)
	}

	if err = json.Unmarshal(resp.BodyBuf.Bytes(), &results); err != nil {
		return results.Spaces, resp, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	return results.Spaces, resp, nil
}

// All lists all Spaces from the Spaces API.
func (c *SpaceClient) All(ctx context.Context) ([]Space, error) {
	var (
		err  error
		resp *Response
	)

	allResults := make([]Space, 0, defaultListResultSize)
	// No pagination support as yet, but support it for future use

	err = c.all(ctx, newEmptyListOpts(), func(opt listOpts) (*Response, error) {
		var listResults []Space

		listResults, resp, err = c.list(ctx, SpaceListOptions{listOpts: opt})
		if err != nil {
			return resp, fmt.Errorf("client.list failed: %w", err)
		}

		allResults = append(allResults, listResults...)
		if len(allResults) >= resp.PageSize {
			resp.MarkPaginationComplete()
		}
		return resp, err
	})

	if err != nil {
		return allResults, fmt.Errorf("client.all failed: %w", err)
	}

	return allResults, err
}

//nolint:dupl // Allow duplicated code blocks in code paths that may change
func (c *SpaceClient) GetByPath(ctx context.Context, spacePath string) (Space, error) {
	var (
		req    *http.Request
		reqURL *url.URL
		resp   *Response
		err    error
		result Space
	)

	reqURL, err = url.Parse(SpaceAPIBasePath)
	if err != nil {
		return result, fmt.Errorf("cannot parse relative url from basepath (%s): %w", SpaceAPIBasePath, err)
	}

	reqURL.Path = path.Join(reqURL.Path, spacePath)

	req, err = c.NewRequest(ctx, "GET", reqURL.String(), nil)
	if err != nil {
		return result, fmt.Errorf("creating new http request for URL (%s): %w", reqURL.String(), err)
	}

	resp, err = c.Do(ctx, req)
	if err != nil {
		return result, fmt.Errorf("client.do failed: %w", err)
	}

	if err = json.Unmarshal(resp.BodyBuf.Bytes(), &result); err != nil {
		return result, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	return result, nil
}
