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
	PlanAPIBasePath = "/plans"
)

// planAllResponse represents the JSON response object returned from the
// GET /plans endpoint.
//
// It differs from Plan namely in that the AvailableReleases returned is
// a list of string, not Release.
//
// Indeed, it exists to resolve differences between index list response and
// other response structures.
type planAllResponse struct {
	// Represents a machine-readable name for the plan.
	Slug string `json:"slug,omitempty"`
	// Represents the human-readable name of the plan.
	Name string `json:"name,omitempty"`
	// Represents the plan price in cents.
	PriceInCents int64 `json:"price_in_cents,omitempty"`
	// Represents the plan billing interval in months.
	BillingIntervalInMonths int `json:"billing_interval_in_months,omitempty"`
	// Indicates whether the plan is single-tenant or not. A value of false
	// indicates the Cluster will share hardware with other Clusters. Single
	// tenant environments can be reached via the public Internet.
	SingleTenant *bool `json:"single_tenant,omitempty"`
	// Indicates whether the plan is on a publicly addressable network.
	// Private plans provide environments that cannot be reached by the public
	// Internet. A VPC connection will be needed to communicate with a private
	// cluster.
	PrivateNetwork *bool `json:"private_network,omitempty"`
	// A collection of search release slugs available for the plan. Additional
	// information about a release can be retrieved from the Releases API.
	AvailableReleases []string `json:"available_releases"`
	AvailableSpaces   []string `json:"available_spaces"`

	// A URI to retrieve more information about this Plan.
	URI string `json:"uri,omitempty"`
}

type planAllResponseList struct {
	Plans []planAllResponse `json:"plans"`
}

type planAllResponseConverter struct{}

// Convert copies a single planAllResponse into a Plan,
// transforming types as needed.
func (c *planAllResponseConverter) Convert(source planAllResponse) Plan {
	plan := Plan{
		AvailableReleases: make([]Release, len(source.AvailableReleases)),
		AvailableSpaces:   make([]Space, len(source.AvailableSpaces)),
	}
	plan.Slug = source.Slug
	plan.Name = source.Name
	plan.PriceInCents = source.PriceInCents
	plan.BillingIntervalInMonths = source.BillingIntervalInMonths
	plan.SingleTenant = source.SingleTenant
	plan.PrivateNetwork = source.PrivateNetwork
	for i, release := range source.AvailableReleases {
		plan.AvailableReleases[i] = Release{Slug: release}
	}
	for i, space := range source.AvailableSpaces {
		plan.AvailableSpaces[i] = Space{Path: space}
	}
	plan.URI = source.URI

	return plan
}

// ConvertItems converts a slice of planAllResponse into a slice of Plan
// by way of the planAllResponseConverter.ConvertItems method.
func (c *planAllResponseConverter) ConvertItems(source []planAllResponse) []Plan {
	var planList []Plan
	if source != nil {
		planList = make([]Plan, len(source))
		for i := range source {
			planList[i] = c.Convert(source[i])
		}
	}
	return planList
}

// Plan represents a subscription plan.
type Plan struct {
	// Represents a machine-readable name for the plan.
	Slug string `json:"slug"`
	// Represents the human-readable name of the plan.
	Name string `json:"name,omitempty"`
	// Represents the plan price in cents.
	PriceInCents int64 `json:"price_in_cents,omitempty"`
	// Represents the plan billing interval in months.
	BillingIntervalInMonths int `json:"billing_interval_months,omitempty"`
	// Indicates whether the plan is single-tenant or not. A value of false
	// indicates the Cluster will share hardware with other Clusters. Single
	// tenant environments can be reached via the public Internet.
	SingleTenant *bool `json:"single_tenant,omitempty"`
	// Indicates whether the plan is on a publicly addressable network.
	// Private plans provide environments that cannot be reached by the public
	// Internet. A VPC connection will be needed to communicate with a private
	// cluster.
	PrivateNetwork *bool `json:"private_network,omitempty"`
	// A collection of search release slugs available for the plan. Additional
	// information about a release can be retrieved from the Releases API.
	AvailableReleases []Release `json:"available_releases"`
	AvailableSpaces   []Space   `json:"available_spaces"`

	// A URI to retrieve more information about this Plan.
	URI string `json:"uri,omitempty"`
}

func (p *Plan) UnmarshalJSON(data []byte) error {
	intermediary := planAllResponse{}
	if err := json.Unmarshal(data, &intermediary); err != nil {
		return fmt.Errorf("unmarshaling into intermediary type: %w", err)
	}

	converter := planAllResponseConverter{}
	converted := converter.Convert(intermediary)
	*p = converted

	return nil
}

// PlansResultList is a wrapper around a slice of
// Plans for json unmarshaling.
type PlansResultList struct {
	Plans []Plan `json:"plans"`
}

func (p *PlansResultList) UnmarshalJSON(data []byte) error {
	planAllResponseList := make([]planAllResponse, 0)

	if err := json.Unmarshal(data, &planAllResponseList); err != nil {
		return fmt.Errorf("unmarshaling into planAllResponseList type: %w", err)
	}

	converter := planAllResponseConverter{}
	p.Plans = converter.ConvertItems(planAllResponseList)
	return nil
}

// PlanClient is a client for the Plans API.
type PlanClient struct {
	*Client
}

type planListOptions struct {
	listOpts
}

func (o planListOptions) values() url.Values {
	return o.listOpts.values()
}

// list returns a list of Plans for the page specified,
// by performing a GET request against [spaceAPIBasePath].
//
// Note: Pagination is not currently supported.
func (c *PlanClient) list(ctx context.Context, opt planListOptions) ([]Plan, *Response, error) {
	var (
		req    *http.Request
		reqURL *url.URL
		resp   *Response
		err    error

		results []Plan
	)
	// Let's make some initial capacity to reduce allocations
	intermediaryResults := planAllResponseList{
		Plans: make([]planAllResponse, 0, defaultResponseCapacity),
	}

	reqURL, err = url.Parse(PlanAPIBasePath)
	if err != nil {
		return results, nil, fmt.Errorf("cannot parse relative url from basepath (%s): %w", PlanAPIBasePath, err)
	}

	// Conditionally set options if we received any
	if !reflect.ValueOf(opt).IsZero() {
		reqURL.RawQuery = opt.values().Encode()
	}

	req, err = c.NewRequest(ctx, "GET", reqURL.String(), nil)
	if err != nil {
		return results, nil, fmt.Errorf("creating new http request for URL (%s): %w", reqURL.String(), err)
	}

	resp, err = c.Do(ctx, req)
	if err != nil {
		return results, resp, fmt.Errorf("client.do failed: %w", err)
	}

	if err = json.Unmarshal(resp.BodyBuf.Bytes(), &intermediaryResults); err != nil {
		return results, resp, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	converter := planAllResponseConverter{}
	results = converter.ConvertItems(intermediaryResults.Plans)

	return results, resp, nil
}

// All lists all Plans from the Plans API.
func (c *PlanClient) All(ctx context.Context) ([]Plan, error) {
	var (
		err  error
		resp *Response
	)

	allResults := make([]Plan, 0, defaultListResultSize)
	// No pagination support as yet, but support it for future use

	err = c.all(ctx, newEmptyListOpts(), func(opt listOpts) (*Response, error) {
		var listResults []Plan

		listResults, resp, err = c.list(ctx, planListOptions{listOpts: opt})
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

// GetBySlug gets a Plan from the Plans API by its slug.
//
//nolint:dupl // Allow duplicated code blocks in code paths that may change
func (c *PlanClient) GetBySlug(ctx context.Context, slug string) (Plan, error) {
	var (
		req    *http.Request
		reqURL *url.URL
		resp   *Response
		err    error
		result Plan
	)

	reqURL, err = url.Parse(PlanAPIBasePath)
	if err != nil {
		return result, fmt.Errorf("cannot parse relative url from basepath (%s): %w", PlanAPIBasePath, err)
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
