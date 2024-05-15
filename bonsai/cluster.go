package bonsai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"reflect"

	"github.com/google/go-querystring/query"
)

const (
	ClusterAPIBasePath = "/clusters"
)

// ClusterStats holds *some* statistics about the cluster.
//
// This attribute should not be used for real-time monitoring!
// Stats are updated every 10-15 minutes. To monitor real-time metrics, monitor
// your cluster directly, via the Index Stats API.
type ClusterStats struct {
	// Number of documents in the index.
	Docs int64 `json:"docs,omitempty"`
	// Number of shards the cluster is using.
	ShardsUsed int64 `json:"shards_used,omitempty"`
	// Number of bytes the cluster is using on-disk.
	DataBytesUsed int64 `json:"data_bytes_used,omitempty"`
}

// ClusterAccess holds information about connecting to the cluster.
type ClusterAccess struct {
	// Host name of the cluster.
	Host string `json:"host"`
	// HTTP Port the cluster is running on.
	Port int `json:"port"`
	// HTTP Scheme needed to access the cluster. Default: "https".
	Scheme string `json:"scheme"`

	// User holds the username to access the cluster with.
	// Only shown once, during cluster creation.
	Username string `json:"user,omitempty"`
	// Pass holds the password to access the cluster with.
	// Only shown once, during cluster creation.
	Password string `json:"pass,omitempty"`
	// URL is the Cluster endpoint for access.
	// Only shown once, during cluster creation.
	URL string `json:"url,omitempty"`
}

// ClusterState represents the current state of the cluster, indicating what
// the cluster is doing at any given moment.
type ClusterState string

const (
	ClusterStateDeprovisioned  ClusterState = "DEPROVISIONED"
	ClusterStateDeprovisioning ClusterState = "DEPROVISIONING"
	ClusterStateDisabled       ClusterState = "DISABLED"
	ClusterStateMaintenance    ClusterState = "MAINTENANCE"
	ClusterStateProvisioned    ClusterState = "PROVISIONED"
	ClusterStateProvisioning   ClusterState = "PROVISIONING"
	ClusterStateReadOnly       ClusterState = "READONLY"
	ClusterStateUpdatingPlan   ClusterState = "UPDATING PLAN"
)

// Cluster represents a single cluster on your account.
type Cluster struct {
	// Slug represents a unique, machine-readable name for the cluster.
	// A cluster slug is based its name at creation, to which a random integer
	// is concatenated.
	Slug string `json:"slug"`
	// Name is the human-readable name of the cluster.
	Name string `json:"name"`
	// URI is a link to additional information about this cluster.
	URI string `json:"uri"`

	// Plan holds some information about the cluster's current subscription plan.
	Plan Plan `json:"plan"`

	// Release holds some information about the cluster's current release.
	Release Release `json:"release"`

	// Space holds some information about where the cluster is running.
	Space Space `json:"space"`

	// Stats holds a collection of statistics about the cluster.
	Stats ClusterStats `json:"stats"`

	// ClusterAccess holds information about connecting to the cluster.
	Access ClusterAccess `json:"access"`

	// State represents the current state of the cluster. This indicates what
	// the cluster is doing at any given moment.
	State ClusterState `json:"state"`
}

type ClusterResultGetBySlug struct {
	Cluster `json:"cluster"`
}

// ClustersResultList is a wrapper around a slice of
// Clusters for json unmarshaling.
type ClustersResultList struct {
	Clusters []Cluster `json:"clusters"`
}

// ClustersResultCreate is the result response for Create (POST) requests to the
// clusters endpoint.
type ClustersResultCreate struct {
	// Message contains details about the cluster creation request.
	Message string `json:"message"`
	// Monitor holds a URI to the Cluster overview page.
	Monitor string        `json:"monitor"`
	Access  ClusterAccess `json:"access"`
}

// ClustersResultUpdate is the result response for Update (PUT) requests to the
// clusters endpoint.
type ClustersResultUpdate struct {
	// Message contains details about the cluster update request.
	Message string `json:"message"`
	// Monitor holds a URI to the Cluster overview page.
	Monitor string `json:"monitor"`
}

// ClustersResultDestroy is the result response for Destroy (DELETE) requests to the
// clusters endpoint.
type ClustersResultDestroy struct {
	// Message contains details about the cluster destroy request.
	Message string `json:"message"`
	// Monitor holds a URI to the Cluster overview page.
	Monitor string `json:"monitor"`
}

// ClusterClient is a client for the Clusters API.
type ClusterClient struct {
	*Client
}

// Do performs an HTTP request against the API, with any required
// Cluster-specific configuration/limitations - for example, rate limiting.
func (c *ClusterClient) Do(ctx context.Context, req *http.Request) (*Response, error) {
	// Allow non-provisioning Cluster endpoint requests to continue
	if req.Method != http.MethodPost {
		return c.Client.Do(ctx, req)
	}

	// Limit provision requests
	err := c.rateLimiter.provisionLimiter.Wait(ctx)
	if err != nil {
		// Context canceled, timed-out, burst issue, or other rate limit issue;
		// let the callers handle it.
		return nil, fmt.Errorf("failed while awaiting execution per rate-limit: %w", err)
	}
	return c.Client.Do(ctx, req)
}

type ClusterAllOpts struct {
	// Optional. A query string for filtering matching clusters.
	// This currently works on name.
	Query string `url:"q,omitempty"`
	// Optional. A string which will constrain results to parent or child
	// cluster. Valid values are: "parent", "child"
	Tenancy string `url:"tenancy,omitempty"`
	// Optional. A string representing the account, region, space, or cluster
	// path where the cluster is located. You can get a list of available spaces
	// with the [bonsai.SpaceClient] API. Space path prefixes work here, so you
	// can find all clusters in a given region for a given cloud.
	Location string `url:"location,omitempty"`
}

type ClusterCreateOpts struct {
	// Required. A String representing the name for the new cluster.
	Name string `json:"name"`
	// The slug of the Plan that the new cluster will be configured for.
	// Use the [PlanClient.All] method to view a list of all Plans available.
	Plan string `json:"plan,omitempty"`
	// The slug of the Space where the new cluster should be deployed to.
	// Use the [SpaceClient.All] method to view a list of all Spaces.
	Space string `json:"space,omitempty"`
	// The Search Service Release that the new cluster will use.
	// Use the [ReleaseClient.All] method to view a list of all Spaces.
	Release string `json:"release,omitempty"`
}

func (o ClusterCreateOpts) Valid() error {
	if o.Name == "" {
		return errors.New("name can't be empty")
	}
	return nil
}

type ClusterUpdateOpts struct {
	// Required. A String representing the name for the new cluster.
	Name string `json:"name"`
	// Required. The slug of the Plan that the new cluster will be configured for.
	// Use the [PlanClient.All] method to view a list of all Plans available.
	Plan string `json:"plan,omitempty"`
}

func (o ClusterUpdateOpts) Valid() error {
	if o.Name == "" {
		return errors.New("name can't be empty")
	}
	return nil
}

type clusterListOpts struct {
	listOpts
	ClusterAllOpts
}

func (o clusterListOpts) values() (url.Values, error) {
	queryValues := o.listOpts.values()

	clusterValues, err := query.Values(o.ClusterAllOpts)
	if err != nil {
		return nil, fmt.Errorf("error parsing cluster list options: %w", err)
	}

	for k, v := range clusterValues {
		queryValues[k] = v
	}

	return queryValues, nil
}

// list returns a list of Clusters for the page specified,
// by performing a GET request against [spaceAPIBasePath].
//
// Note: Pagination is not currently supported.
func (c *ClusterClient) list(ctx context.Context, opt clusterListOpts) (
	[]Cluster,
	*Response,
	error,
) {
	var (
		req    *http.Request
		reqURL *url.URL
		resp   *Response
		err    error
	)
	// Let's make some initial capacity to reduce allocations
	results := ClustersResultList{
		Clusters: make([]Cluster, 0, defaultResponseCapacity),
	}

	reqURL, err = url.Parse(ClusterAPIBasePath)
	if err != nil {
		return results.Clusters, nil, fmt.Errorf("cannot parse relative url from basepath (%s): %w", ClusterAPIBasePath, err)
	}

	// Conditionally set options if we received any
	if !reflect.ValueOf(opt).IsZero() {
		var optVals url.Values

		optVals, err = opt.values()
		if err != nil {
			return results.Clusters, nil, fmt.Errorf("failed to get values from opt (%+v): %w", opt, err)
		}

		reqURL.RawQuery = optVals.Encode()
	}

	req, err = c.NewRequest(ctx, "GET", reqURL.String(), nil)
	if err != nil {
		return results.Clusters, nil, fmt.Errorf("creating new http request for URL (%s): %w", reqURL.String(), err)
	}

	resp, err = c.Do(ctx, req)
	if err != nil {
		return results.Clusters, resp, fmt.Errorf("client.do failed: %w", err)
	}

	if err = json.Unmarshal(resp.BodyBuf.Bytes(), &results); err != nil {
		return results.Clusters, resp, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	return results.Clusters, resp, nil
}

// All lists all active clusters on your account.
func (c *ClusterClient) All(ctx context.Context) ([]Cluster, error) {
	var (
		err  error
		resp *Response
	)

	allResults := make([]Cluster, 0, defaultListResultSize)
	// No pagination support as yet, but support it for future use

	err = c.all(ctx, newEmptyListOpts(), func(opt listOpts) (*Response, error) {
		var listResults []Cluster

		listResults, resp, err = c.list(ctx, clusterListOpts{listOpts: opt})
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

// GetBySlug gets a Cluster from the Clusters API by its slug.
func (c *ClusterClient) GetBySlug(ctx context.Context, slug string) (Cluster, error) {
	var (
		req                *http.Request
		reqURL             *url.URL
		resp               *Response
		err                error
		result             Cluster
		intermediaryResult ClusterResultGetBySlug
	)

	reqURL, err = url.Parse(ClusterAPIBasePath)
	if err != nil {
		return result, fmt.Errorf("cannot parse relative url from basepath (%s): %w", ClusterAPIBasePath, err)
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

	if err = json.Unmarshal(resp.BodyBuf.Bytes(), &intermediaryResult); err != nil {
		return result, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	return intermediaryResult.Cluster, nil
}

// Create requests a new Cluster to be created.
//

func (c *ClusterClient) Create(ctx context.Context, opt ClusterCreateOpts) (
	ClustersResultCreate,
	error,
) {
	var (
		req     *http.Request
		reqURL  *url.URL
		reqBody []byte
		resp    *Response
		err     error
	)
	// Let's make some initial capacity to reduce allocations
	result := ClustersResultCreate{}

	reqURL, err = url.Parse(ClusterAPIBasePath)
	if err != nil {
		return result, fmt.Errorf("cannot parse relative url from basepath (%s): %w", ClusterAPIBasePath, err)
	}

	if err = opt.Valid(); err != nil {
		return result, fmt.Errorf("invalid create options (%v): %w", opt, err)
	}

	reqBody, err = json.Marshal(opt)
	if err != nil {
		return result, fmt.Errorf("failed to marshal options (%v): %w", opt, err)
	}

	req, err = c.NewRequest(ctx, "POST", reqURL.String(), bytes.NewReader(reqBody))
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

// Update requests a new Cluster be updated.
//

func (c *ClusterClient) Update(ctx context.Context, slug string, opt ClusterUpdateOpts) (
	ClustersResultUpdate,
	error,
) {
	var (
		req     *http.Request
		reqURL  *url.URL
		reqBody []byte
		resp    *Response
		err     error
	)
	// Let's make some initial capacity to reduce allocations
	result := ClustersResultUpdate{}

	reqURL, err = url.Parse(ClusterAPIBasePath)
	if err != nil {
		return result, fmt.Errorf("cannot parse relative url from basepath (%s): %w", ClusterAPIBasePath, err)
	}
	reqURL.Path = path.Join(reqURL.Path, slug)

	if err = opt.Valid(); err != nil {
		return result, fmt.Errorf("invalid create options (%v): %w", opt, err)
	}

	reqBody, err = json.Marshal(opt)
	if err != nil {
		return result, fmt.Errorf("failed to marshal options (%v): %w", opt, err)
	}

	req, err = c.NewRequest(ctx, "PUT", reqURL.String(), bytes.NewReader(reqBody))
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

// Destroy triggers the deprovisioning of the cluster associated with the slug.
//
//nolint:dupl // Allow duplicated code blocks in code paths that may change
func (c *ClusterClient) Destroy(ctx context.Context, slug string) (ClustersResultDestroy, error) {
	var (
		req    *http.Request
		reqURL *url.URL
		resp   *Response
		err    error
		result ClustersResultDestroy
	)

	reqURL, err = url.Parse(ClusterAPIBasePath)
	if err != nil {
		return result, fmt.Errorf("cannot parse relative url from basepath (%s): %w", ClusterAPIBasePath, err)
	}

	reqURL.Path = path.Join(reqURL.Path, slug)

	req, err = c.NewRequest(ctx, "DELETE", reqURL.String(), nil)
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
