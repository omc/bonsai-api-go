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

const ReleaseAPIBasePath = "/releases"

// Release is a version of Elasticsearch available to the account.
type Release struct {
	// The name for the release.
	Name string `json:"name,omitempty"`
	// The machine-readable name for the deployment.
	Slug string `json:"slug,omitempty"`
	// The service type of the deployment - for example, "elasticsearch".
	ServiceType string `json:"service_type,omitempty"`
	// The version of the release.
	Version string `json:"version,omitempty"`
	// Whether the release is available on multitenant deployments.
	MultiTenant *bool  `json:"multitenant,omitempty"`

	// A URI to retrieve more information about this Release.
	URI string `json:"uri,omitempty"`
	// PackageName is the package name of the release.
	PackageName string `json:"package_name,omitempty"`
}

// ReleasesResultList is a wrapper around a slice of
// Releases for json unmarshaling.
type ReleasesResultList struct {
	Releases []Release `json:"releases,omitempty"`
}

// ReleaseClient is a client for the Releases API.
type ReleaseClient struct {
	*Client
}

type releaseListOptions struct {
	listOpts
}

func (o releaseListOptions) values() url.Values {
	return o.listOpts.values()
}

// list returns a list of Releases for the page specified,
// by performing a GET request against [spaceAPIBasePath].
//
// Note: Pagination is not currently supported.
func (c *ReleaseClient) list(ctx context.Context, opt releaseListOptions) ([]Release, *Response, error) {
	var (
		req    *http.Request
		reqURL *url.URL
		resp   *Response
		err    error
	)
	// Let's make some initial capacity to reduce allocations
	results := ReleasesResultList{
		Releases: make([]Release, 0, defaultResponseCapacity),
	}

	reqURL, err = url.Parse(ReleaseAPIBasePath)
	if err != nil {
		return results.Releases, nil, fmt.Errorf("cannot parse relative url from basepath (%s): %w", ReleaseAPIBasePath, err)
	}

	// Conditionally set options if we received any
	if !reflect.ValueOf(opt).IsZero() {
		reqURL.RawQuery = opt.values().Encode()
	}

	req, err = c.NewRequest(ctx, "GET", reqURL.String(), nil)
	if err != nil {
		return results.Releases, nil, fmt.Errorf("creating new http request for URL (%s): %w", reqURL.String(), err)
	}

	resp, err = c.Do(ctx, req)
	if err != nil {
		return results.Releases, resp, fmt.Errorf("client.do failed: %w", err)
	}

	if err = json.Unmarshal(resp.BodyBuf.Bytes(), &results); err != nil {
		return results.Releases, resp, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	return results.Releases, resp, nil
}

// All lists all Releases from the Releases API.
func (c *ReleaseClient) All(ctx context.Context) ([]Release, error) {
	var (
		err  error
		resp *Response
	)

	allResults := make([]Release, 0, defaultListResultSize)
	// No pagination support as yet, but support it for future use

	err = c.all(ctx, newEmptyListOpts(), func(opt listOpts) (*Response, error) {
		var listResults []Release

		listResults, resp, err = c.list(ctx, releaseListOptions{listOpts: opt})
		if err != nil {
			return resp, fmt.Errorf("client.list failed: %w", err)
		}

		allResults = append(allResults, listResults...)
		if len(allResults) >= resp.TotalRecords {
			resp.MarkPaginationComplete()
		}
		return resp, err
	})

	if err != nil {
		return allResults, fmt.Errorf("client.all failed: %w", err)
	}

	return allResults, err
}

// GetBySlug gets a Release from the Releases API by its slug.
//
//nolint:dupl // Allow duplicated code blocks in code paths that may change
func (c *ReleaseClient) GetBySlug(ctx context.Context, slug string) (Release, error) {
	var (
		req    *http.Request
		reqURL *url.URL
		resp   *Response
		err    error
		result Release
	)

	reqURL, err = url.Parse(ReleaseAPIBasePath)
	if err != nil {
		return result, fmt.Errorf("cannot parse relative url from basepath (%s): %w", ReleaseAPIBasePath, err)
	}

	reqURL.Path = path.Join(reqURL.Path, slug)

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
