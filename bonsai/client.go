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
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/http/httpguts"
	"golang.org/x/time/rate"
)

// Client representation configuration.
const (
	// Version reflects this API Client's version.
	Version = "1.0.0"
	// BaseEndpoint is the target API URL base location.
	BaseEndpoint = "https://api.bonsai.io"
	// UserAgent is the internally used value for the User-Agent header
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

// Values returns the listOpts as URL values.
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

type Token struct {
	string
}

func (t Token) Empty() bool {
	return t.string == ""
}

func (t Token) NotEmpty() bool {
	return !t.Empty()
}

func NewToken(token string) (Token, error) {
	t := Token{token}
	if ok := t.validHTTPValue(); !ok {
		return Token{}, errors.New("invalid token")
	}
	return t, nil
}

func (t Token) validHTTPValue() bool {
	return httpguts.ValidHeaderFieldValue(t.string)
}

// ClientOption is a functional option, used to configure Client.
type ClientOption func(*Client)

// WithEndpoint configures a Client to use the specified API endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(c *Client) {
		c.endpoint = strings.TrimRight(endpoint, "/")
	}
}

// WithToken configures a Client to use the specified token for authentication.
func WithToken(token Token) ClientOption {
	return func(c *Client) {
		c.token = token
	}
}

// WithApplication configures the client to represent itself as
// a particular Application by modifying the User-Agent header
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

type PaginatedResponse struct {
	PageNumber   int `json:"page_number"`
	PageSize     int `json:"page_size"`
	TotalRecords int `json:"total_records"`
}

type httpResponse = *http.Response
type Response struct {
	httpResponse      `json:"-"`
	BodyBuf           *bytes.Buffer `json:"-"`
	PaginatedResponse `json:"pagination"`
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

	rateLimiter *ClientLimiter
	endpoint    string
	token       Token
	userAgent   string
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

	return client
}

func (c *Client) UserAgent() string {
	return c.userAgent
}

// NewRequest creates an HTTP request against the API. The returned request
// is assigned with ctx and has all necessary headers set (auth, user agent, etc.).
func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	reqURL := c.endpoint + path
	req, err := http.NewRequest(method, reqURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)

	if c.token.NotEmpty() {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

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
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
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

	// Extract the pagination details
	if httpResp.Header.Get("Content-Type") == HTTPContentTypeJSON {
		err = json.Unmarshal(resp.BodyBuf.Bytes(), &resp)
		if err != nil {
			return resp, fmt.Errorf("error unmarshaling response body for pagination: %w", err)
		}
	}

	if resp.StatusCode >= http.StatusBadRequest {
		respErr := ResponseError{}
		if err = json.Unmarshal(resp.BodyBuf.Bytes(), &respErr); err != nil {
			return resp, fmt.Errorf("unmarshalling error response: %w", err)
		}
		return resp, respErr
	}

	return resp, err
}

// all loops through the next page pagination results until empty
// it allows the caller to pass a func (typically a closure) to collect
// results.
func (c *Client) all(ctx context.Context, f func(int) (*Response, error)) error {
	var (
		page = 1
	)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp, err := f(page)
			if err != nil {
				return err
			}

			// The caller is responsible for determining whether or not we've exhausted
			// retries.
			if reflect.ValueOf(resp.PaginatedResponse).IsZero() || resp.PageNumber <= 0 {
				return nil
			}
			// We should be fine with a straight increment, but let's play it safe
			page = resp.PageNumber + 1
		}
	}
}
