package bonsai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/time/rate"
)

// ClientImplTestSuite is responsible for testing internal facing items/behavior
// that isn't part of the exposed interface, but is hard to test via that interface.
// Things like the default HTTP Client's rate-limiter (unexposed), and other implementation
// details fall under this umbrella - these test cases should be few.
type ClientImplTestSuite struct {
	// Assertions embedded here allows all tests to reach through the suite to access assertion methods
	*require.Assertions
	// Suite is the testify/suite used for all HTTP request tests
	suite.Suite

	// serveMux is the request multiplexer used for tests
	serveMux *chi.Mux
	// server is the testing server on some local port
	server *httptest.Server
	// client allows each test to have a reachable *Client for testing
	client *Client
}

func (s *ClientImplTestSuite) SetupSuite() {
	// Configure http client and other miscellany
	s.serveMux = chi.NewRouter()
	s.server = httptest.NewServer(s.serveMux)

	user, err := NewAccessKey("TestUser")
	if err != nil {
		log.Fatal(fmt.Errorf("invalid user received: %w", err))
	}

	password, err := NewAccessToken("TestToken")
	if err != nil {
		log.Fatal(fmt.Errorf("invalid token/password received: %w", err))
	}

	s.client = NewClient(
		WithEndpoint(s.server.URL),
		WithCredentialPair(
			CredentialPair{
				AccessKey:   user,
				AccessToken: password,
			},
		),
	)

	// configure testify
	s.Assertions = require.New(s.T())
}

func (s *ClientImplTestSuite) TestClientDefaultRateLimit() {
	c := NewClient()
	s.Equal(DefaultClientBurstAllowance, c.rateLimiter.Burst())
	s.InEpsilon(float64(rate.Every(DefaultClientBurstDuration)), float64(c.rateLimiter.Limit()), Float64Epsilon)
}

func (s *ClientImplTestSuite) TestListOptsValues() {
	testCases := []struct {
		name     string
		received listOpts
		expect   string
	}{
		{
			name: "with populated values",
			received: listOpts{
				Page: 3,
				Size: 100,
			},
			expect: "page=3&size=100",
		},
		{
			name:     "with empty values",
			received: listOpts{},
			expect:   "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.Equal(tc.received.values().Encode(), tc.expect)
		})
	}
}

func (s *ClientImplTestSuite) TestClientAll() {
	const expectedPageCount = 4
	var (
		ctx          = context.Background()
		expectedPage = 1
	)

	s.serveMux.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HTTPHeaderContentType, HTTPContentTypeJSON)

		respBody, _ := NewResponse()
		respBody.PaginatedResponse = PaginatedResponse{
			PageNumber:   3,
			PageSize:     1,
			TotalRecords: 3,
		}

		switch page := r.URL.Query().Get("page"); page {
		case "", "1":
			respBody.PaginatedResponse.PageNumber = 1
		case "2":
			respBody.PaginatedResponse.PageNumber = 2
		case "3":
			respBody.PaginatedResponse.PageNumber = 3
		default:
			s.FailNowf("invalid page parameter", "page parameter: %v", page)
		}

		err := json.NewEncoder(w).Encode(respBody)
		s.NoError(err, "encode response body")
	})

	// The caller must track results against expected count
	// A reminder to the reader: this is the caller.
	var resultCount = 0
	err := s.client.all(context.Background(), newEmptyListOpts(), func(opt listOpts) (*Response, error) {
		reqPath := "/"

		if opt.Valid() {
			s.Equalf(expectedPage, opt.Page, "expected page number (%d) matches actual (%d)", expectedPage, opt.Page)
			reqPath = fmt.Sprintf("%s?page=%d&size=1", reqPath, opt.Page)
		}

		req, err := s.client.NewRequest(ctx, "GET", reqPath, nil)
		s.NoError(err, "new request for path")

		resp, err := s.client.Do(context.Background(), req)
		s.NoError(err, "do request")

		expectedPage++
		// A reference of how these funcs should handle this;
		// recall, the response may be shorter than max.
		//
		// Ideally, this count wouldn't be derived from PageSize,
		// but rather, from the total count of discovered items
		// unmarshaled.
		resultCount += max(resp.PageSize, 0)

		if resultCount >= resp.TotalRecords {
			resp.MarkPaginationComplete()
		}
		return resp, err
	})
	s.NoError(err, "client.all call")

	s.Equalf(
		expectedPage,
		expectedPageCount,
		"expected page visit count (%d) matches actual visit count (%d)",
		expectedPageCount-1,
		expectedPage-1,
	)
}

func TestClientImplTestSuite(t *testing.T) {
	suite.Run(t, new(ClientImplTestSuite))
}
