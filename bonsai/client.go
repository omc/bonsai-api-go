package bonsai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/time/rate"
)

// Client representation configuration.
const (
	// Version reflects this API Client's version.
	Version = "v2.2.0"
	// BaseEndpoint is the target API URL base location.
	BaseEndpoint = "https://api.bonsai.io"
	// UserAgent is the internally used value for the AccessKey-Agent header
	// in all outgoing HTTP requests.
	UserAgent = "bonsai-api-go/" + Version
)

// Client rate limiter configuration.
const (
	// DefaultClientBurstAllowance is the default Bonsai API request burst
	// allowance.
	DefaultClientBurstAllowance = 60
	// DefaultClientBurstDuration is the default interval for a token
	// bucket of size DefaultClientBurstAllowance to be refilled.
	DefaultClientBurstDuration = 1 * time.Minute
	// ProvisionClientBurstAllowance is the default Bonsai API request burst allowance.
	ProvisionClientBurstAllowance = 5
	// ProvisionClientBurstDuration is the default interval for a token bucket
	// of size ProvisionClientBurstAllowance to be refilled.
	ProvisionClientBurstDuration = 1 * time.Minute
)

// Common API Response headers.
const (
	// HeaderRetryAfter holds the number of seconds to delay before making the next request.
	// ref: https://bonsai.io/docs/api-error-429-too-many-requests
	HeaderRetryAfter = "Retry-After"
)

// HTTP Content Types and related Header.
const (
	HTTPHeaderContentType        = "Content-Type"
	HTTPContentTypeJSON   string = "application/json"
)

// Magic numbers used to limit allocations, etc.
const (
	defaultResponseCapacity = 8
	defaultListResultSize   = 100
)

// HTTP Status Response Errors.
var (
	ErrHTTPStatusNotFound            = errors.New("not found")
	ErrHTTPStatusForbidden           = errors.New("forbidden")
	ErrHTTPStatusPaymentRequired     = errors.New("payment required")
	ErrHTTPStatusUnprocessableEntity = errors.New("unprocessable entity")
	ErrHTTPStatusUnauthorized        = errors.New("unauthorized")
	ErrHTTPStatusTooManyRequests     = errors.New("too many requests")
)

var contentTypeRegexp = regexp.MustCompile(fmt.Sprintf("^%s.*", HTTPContentTypeJSON))

// ResponseError captures API response errors
// returned as JSON in supported scenarios.
//
// ref: https://bonsai.io/docs/introduction-to-the-api
type ResponseError struct {
	Errors []string `json:"errors"`
	Status int      `json:"status"`
}

// Error represents ResponseError, which may have multiple Errors
// as a string.
//
// The community is as yet undecided on a great way to handle this
// ref: https://github.com/golang/go/issues/47811
func (r ResponseError) Error() string {
	return fmt.Sprintf("%v (%d)", r.Errors, r.Status)
}

func (r ResponseError) Is(target error) bool {
	switch r.Status {
	case http.StatusUnauthorized:
		return target == ErrHTTPStatusUnauthorized
	case http.StatusNotFound:
		return target == ErrHTTPStatusNotFound
	case http.StatusForbidden:
		return target == ErrHTTPStatusForbidden
	case http.StatusPaymentRequired:
		return target == ErrHTTPStatusPaymentRequired
	case http.StatusUnprocessableEntity:
		return target == ErrHTTPStatusUnprocessableEntity
	case http.StatusTooManyRequests:
		return target == ErrHTTPStatusTooManyRequests
	}

	return false
}

// listOpts specifies options for listing resources.
// ref: https://bonsai.io/docs/api-result-pagination
type listOpts struct {
	Page int // Page number, starting at 1
	Size int // Size of each page, with a max of 100
}

// newListOpts creates a new listOpts with default values per the API docs.
//
//nolint:unused // will be used for clusters endpoint
func newDefaultListOpts() listOpts {
	return listOpts{
		Page: 1,
		Size: defaultListResultSize,
	}
}

// newEmptyListOpts returns an empty list opts,
// to make it easy for readers to immediately see that there are no options
// being passed, rather than seeing a struct be initialized in-line.
func newEmptyListOpts() listOpts {
	return listOpts{}
}

// values returns the listOpts as URL values.
func (l listOpts) values() url.Values {
	vals := url.Values{}
	if l.Page > 0 {
		vals.Add("page", strconv.Itoa(l.Page))
	}
	if l.Size > 0 {
		vals.Add("size", strconv.Itoa(l.Size))
	}
	return vals
}

func (l listOpts) IsZero() bool {
	return l.Page == 0 && l.Size == 0
}

func (l listOpts) Valid() bool {
	return !l.IsZero()
}

type Application struct {
	Name    string
	Version string
}

func (app Application) String() string {
	switch {
	case app.Name != "" && app.Version != "":
		return app.Name + "/" + app.Version
	case app.Name != "" && app.Version == "":
		return app.Name
	default:
		return ""
	}
}

type Credential string

type AccessKey Credential

// NewAccessKey is a convenience method for verifying
// that access keys intended to be used with the API are valid HTTP header values.
func NewAccessKey(user string) (AccessKey, error) {
	if ok := Credential(user).validHTTPValue(); !ok {
		return AccessKey(""), errors.New("invalid user")
	}
	return AccessKey(user), nil
}

type AccessToken Credential

// NewAccessToken is a convenience method for verifying
// that access tokens intended to be used with the API are valid HTTP
// header values.
func NewAccessToken(password string) (AccessToken, error) {
	if ok := Credential(password).validHTTPValue(); !ok {
		return AccessToken(""), errors.New("invalid password")
	}
	return AccessToken(password), nil
}

func (c Credential) Empty() bool {
	return c == ""
}

func (c Credential) NotEmpty() bool {
	return !c.Empty()
}

func (c Credential) validHTTPValue() bool {
	return httpguts.ValidHeaderFieldValue(string(c))
}

type CredentialPair struct {
	AccessKey
	AccessToken
}

func (c CredentialPair) Empty() bool {
	return reflect.ValueOf(c).IsZero()
}

func (c CredentialPair) NotEmpty() bool {
	return !c.Empty()
}

// ClientOption is a functional option, used to configure Client.
type ClientOption func(*Client)

// WithEndpoint configures a Client to use the specified API endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(c *Client) {
		c.endpoint = strings.TrimRight(endpoint, "/")
	}
}

// WithCredentialPair configures a Client to use
// the specified username for Basic authorization.
func WithCredentialPair(pair CredentialPair) ClientOption {
	return func(c *Client) {
		c.credentialPair = pair
	}
}

// WithApplication configures the client to represent itself as
// a particular Application by modifying the AccessKey-Agent header
// sent in all requests.
func WithApplication(app Application) ClientOption {
	return func(c *Client) {
		c.userAgent = app.String()
		if c.userAgent == "" {
			c.userAgent = UserAgent
		} else {
			c.userAgent += " " + UserAgent
		}
	}
}

// WithDefaultRateLimit configures the default rate limit for client requests.
func WithDefaultRateLimit(l *rate.Limiter) ClientOption {
	return func(c *Client) {
		c.rateLimiter.limiter = l
	}
}

// WithProvisionRateLimit configures the rate limit for client requests to the Provision API.
func WithProvisionRateLimit(l *rate.Limiter) ClientOption {
	return func(c *Client) {
		c.rateLimiter.provisionLimiter = l
	}
}

// WithHTTPTransport configures the Client's HTTP Transport, such that
// "the mechanism by which individual HTTP requests are made" can be
// overridden.
func WithHTTPTransport(t http.RoundTripper) ClientOption {
	return func(c *Client) {
		c.httpClient.Transport = t
	}
}

type PaginatedResponse struct {
	PageNumber   int `json:"page_number"`
	PageSize     int `json:"page_size"`
	TotalRecords int `json:"total_records"`
}

type httpResponse = *http.Response
type Response struct {
	httpResponse      `json:"-"`
	BodyBuf           bytes.Buffer `json:"-"`
	PaginatedResponse `json:"pagination"`
}

func (r *Response) isJSON() bool {
	return contentTypeRegexp.MatchString(r.Header.Get("Content-Type"))
}

// WithHTTPResponse assigns an *http.Response to a *Response item
// and reads its response body into the *Response.
func (r *Response) WithHTTPResponse(httpResp *http.Response) error {
	r.httpResponse = httpResp

	err := r.readHTTPResponseBody()
	if err != nil {
		return fmt.Errorf("reading response body for error extraction: %w", err)
	}

	return err
}

func (r *Response) MarkPaginationComplete() {
	r.PaginatedResponse = PaginatedResponse{}
}

func (r *Response) readHTTPResponseBody() error {
	var (
		err error
	)

	_, err = r.BodyBuf.ReadFrom(r.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	return nil
}

func extractRetryDelay(r *Response) (int64, error) {
	var (
		retryAfter int64
		err        error
	)
	// We're already blocking on this routine, so sleep inline per the header request.
	if retryAfterStr := r.Header.Get(HeaderRetryAfter); retryAfterStr != "" {
		retryAfter, err = strconv.ParseInt(retryAfterStr, 10, 64)
		if err != nil {
			return retryAfter, fmt.Errorf("error parsing retry-after response: %w", err)
		}
	}
	return retryAfter, nil
}

// NewResponse reserves this function signature, and is
// the recommended way to instantiate a Response, as its behavior
// may change.
func NewResponse() (*Response, error) {
	return &Response{}, nil
}

type limiter = *rate.Limiter
type ClientLimiter struct {
	// limiter is an embedded default rate limiter, but not exposed.
	limiter
	// provisionLimiter is the rate limiter to be used for Provision endpoints
	provisionLimiter *rate.Limiter
}

// Client is the exported client that users interact with.
type Client struct {
	httpClient *http.Client

	rateLimiter    *ClientLimiter
	endpoint       string
	credentialPair CredentialPair
	userAgent      string

	// Clients
	Space   SpaceClient
	Plan    PlanClient
	Release ReleaseClient
	Cluster ClusterClient
}

func NewClient(options ...ClientOption) *Client {
	client := &Client{
		endpoint:   BaseEndpoint,
		httpClient: &http.Client{},
		rateLimiter: &ClientLimiter{
			limiter:          rate.NewLimiter(rate.Every(DefaultClientBurstDuration), DefaultClientBurstAllowance),
			provisionLimiter: rate.NewLimiter(rate.Every(ProvisionClientBurstDuration), ProvisionClientBurstAllowance),
		},
	}

	for _, option := range options {
		option(client)
	}

	// Configure child clients
	client.Space = SpaceClient{client}
	client.Plan = PlanClient{client}
	client.Release = ReleaseClient{client}
	client.Cluster = ClusterClient{client}

	return client
}

// Transport returns the HTTP transport used by the Client to make requests.
func (c *Client) Transport() http.RoundTripper {
	return c.httpClient.Transport
}

// Transport returns the HTTP transport used by the Client to make requests.
func (c *Client) SetTransport(t http.RoundTripper) {
	c.httpClient.Transport = t
}

func (c *Client) UserAgent() string {
	return c.userAgent
}

// NewRequest creates an HTTP request against the API. The returned request
// is assigned with ctx and has all necessary headers set (auth, user agent, etc.).
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	reqURL := c.endpoint + path
	req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("unable to create new request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	if c.credentialPair.NotEmpty() {
		req.SetBasicAuth(
			string(c.credentialPair.AccessKey),
			string(c.credentialPair.AccessToken),
		)

		if _, _, ok := req.BasicAuth(); !ok {
			return nil, errors.New("invalid credentials")
		}
	}

	req.Header.Set("Content-Type", HTTPContentTypeJSON)
	req.Header.Set("Accept", HTTPContentTypeJSON)

	req = req.WithContext(ctx)

	return req, nil
}

// Do performs an HTTP request against the API.
func (c *Client) Do(ctx context.Context, req *http.Request) (*Response, error) {
	reqBuf := new(bytes.Buffer)

	// Capture the original request body
	if req.ContentLength > 0 {
		_, err := reqBuf.ReadFrom(req.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading request body: %w", err)
		}

		err = IoClose(req.Body, err)
		if err != nil {
			return nil, err
		}
	}

	// We only retry in the scenario of http.StatusTooManyRequests (429).
	for {
		respErr := &ResponseError{}
		resp, err := c.doRequest(ctx, req, reqBuf)

		switch {
		case errors.As(err, respErr):
			if reflect.ValueOf(respErr).IsZero() {
				return resp, fmt.Errorf("unknown error occurred with response status %d", resp.StatusCode)
			} else if errors.Is(err, ErrHTTPStatusTooManyRequests) {
				// Block in this routine, if needed.
				var delay int64
				if delay, err = extractRetryDelay(resp); err != nil {
					time.Sleep(time.Duration(delay) * time.Second)
				}
				continue
			}
			return resp, err
		default:
			return resp, err
		}
	}
}

func (c *Client) doRequest(ctx context.Context, req *http.Request, reqBuf *bytes.Buffer) (*Response, error) {
	// Wrap the buffer in a no-op Closer, such that
	// it satisfies the ReadCloser interface
	if req.ContentLength > 0 {
		req.Body = io.NopCloser(reqBuf)
	}

	// Context canceled, timed-out, burst issue, or other rate limit issue;
	// let the callers handle it.
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("failed while awaiting execution per rate-limit: %w", err)
	}

	httpResp, err := c.httpClient.Do(req)
	defer func() { err = IoClose(httpResp.Body, err) }()

	if httpResp == nil {
		return nil, errors.New("received nil http.Response")
	}

	resp, err := NewResponse()
	if err != nil {
		return resp, errors.New("creating new Response")
	}

	err = resp.WithHTTPResponse(httpResp)
	if err != nil {
		return resp, fmt.Errorf("setting http response: %w", err)
	}

	// All error reposes should come with a JSON response per the Error handling
	// section @ https://bonsai.io/docs/introduction-to-the-api.
	//
	// That said, in the scenario that an error *isn't* returned as a JSON body
	// response, it would be jarring to receive a message about an internal
	// unmarshaling attempt, rather than to receive the HTTP Status Error
	if resp.StatusCode >= http.StatusBadRequest {
		respErr := ResponseError{Status: resp.StatusCode}

		if ok := resp.isJSON(); ok {
			// Suppress unmarshaling errors in the event that the response didn't
			// contain a message.
			_ = json.Unmarshal(resp.BodyBuf.Bytes(), &respErr)
		}

		return resp, respErr
	}

	// Extract the pagination details
	if resp.isJSON() {
		err = json.Unmarshal(resp.BodyBuf.Bytes(), &resp)
		if err != nil {
			return resp, fmt.Errorf("error unmarshaling response body for pagination: %w", err)
		}
	}

	return resp, err
}

// all loops through the next page pagination results until empty
// it allows the caller to pass a func (typically a closure) to collect
// results.
func (c *Client) all(ctx context.Context, opt listOpts, f func(opts listOpts) (*Response, error)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp, err := f(opt)
			if err != nil {
				return err
			}

			// The caller is responsible for determining whether we've exhausted
			// retries.
			if reflect.ValueOf(resp.PaginatedResponse).IsZero() || resp.PageNumber <= 0 {
				return nil
			}

			// If the response contains a page number, provide the next call with an
			// incremented page number, and the response page size.
			//
			// Again, the caller must determine whether the total number of results have been delivered.
			if resp.PageNumber > 0 {
				opt = listOpts{
					Page: resp.PageNumber + 1,
					Size: resp.PageSize,
				}
			}
		}
	}
}
