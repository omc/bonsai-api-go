package bonsai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/go-chi/chi/v5"
	"github.com/omc/bonsai-api-go/v1/bonsai"
)

const (
	ResponseErrorHTTPStatusNotFound = `
		{  
			"errors": [    
				"Cluster doesnotexist-1234 not found.",
				"Please review the documentation available at https://docs.bonsai.io",
				"Undefined request."  
			],  
			"status": 404
		}`
)

type ClientTestSuite struct {
	// Assertions embedded here allows all tests to reach through the suite to access assertion methods
	*require.Assertions
	// Suite is the testify/suite used for all HTTP request tests
	suite.Suite

	// serveMux is the request multiplexer used for tests
	serveMux *chi.Mux
	// server is the testing server on some local port
	server *httptest.Server
	// client allows each test to have a reachable *bonsai.Client for testing
	client *bonsai.Client
}

func (s *ClientTestSuite) SetupSuite() {
	// Configure http client and other miscellany
	s.serveMux = chi.NewRouter()
	s.server = httptest.NewServer(s.serveMux)
	token, err := bonsai.NewToken("TestToken")
	if err != nil {
		log.Fatal(fmt.Errorf("invalid token received: %w", err))
	}
	s.client = bonsai.NewClient(
		bonsai.WithEndpoint(s.server.URL),
		bonsai.WithToken(token),
	)

	// configure testify
	s.Assertions = require.New(s.T())
}

func (s *ClientTestSuite) TestResponseErrorUnmarshallJson() {
	testCases := []struct {
		name     string
		received string
		expect   bonsai.ResponseError
	}{
		{
			name: "error example from docs site",
			received: `
				{
					"errors": [
						"This request has failed authentication. ` +
				`Please read the docs or email us at support@bonsai.io."
					],
					"status": 401
				}
			`,
			expect: bonsai.ResponseError{
				Errors: []string{
					"This request has failed authentication. Please read the docs or email us at support@bonsai.io.",
				},
				Status: 401,
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			respErr := bonsai.ResponseError{}
			err := json.Unmarshal([]byte(tc.received), &respErr)
			s.NoError(err)
			s.Equal(tc.expect, respErr)
		})
	}
}

func (s *ClientTestSuite) TestClientResponseError() {
	const p = "/clusters/doesnotexist-1234"

	// Configure Servemux to serve the error response at this path
	s.serveMux.HandleFunc(p, func(w http.ResponseWriter, _ *http.Request) {
		var err error

		w.Header().Set("Content-Type", bonsai.HTTPContentTypeJSON)
		w.WriteHeader(http.StatusNotFound)

		respErr := &bonsai.ResponseError{}
		err = json.Unmarshal([]byte(ResponseErrorHTTPStatusNotFound), respErr)
		s.NoError(err, "successfully unmarshals json into bonsaiResponseError")

		err = json.NewEncoder(w).Encode(respErr)
		s.NoError(err, "encodes http response into ResponseError")
	})

	req, err := s.client.NewRequest(context.Background(), "GET", p, nil)
	s.NoError(err, "request creation returns no error")

	resp, err := s.client.Do(context.Background(), req)
	s.Error(err, "Client.Do returns an error")

	s.Equal(http.StatusNotFound, resp.StatusCode)
	s.ErrorAs(err, &bonsai.ResponseError{}, "Client.Do error response type is of ResponseError")
	s.ErrorIs(err, bonsai.ErrHTTPStatusNotFound, "ResponseError is comparable to bonsai.ErrorHttpResponseStatus")
}

func (s *ClientTestSuite) TestClientResponseWithPagination() {
	s.serveMux.HandleFunc("/clusters", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("RateLimit-Limit", "1000")
		w.Header().Set("RateLimit-Remaining", "999")
		w.Header().Set("RateLimit-Reset", "1511954577")
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, `
			{
				"foo": "bar",
				"pagination": {
					"page_number": 1,
					"page_size": 20,
					"total_records": 255
				}
			}
		`)
		s.NoError(err, "writes json response into response writer")
	})

	req, err := s.client.NewRequest(context.Background(), "GET", "/clusters", nil)
	s.NoError(err, "request creation returns no error")

	resp, err := s.client.Do(context.Background(), req)
	s.NoError(err, "Client.Do succeeds")

	s.Equal(1, resp.PaginatedResponse.PageNumber)
	s.Equal(20, resp.PaginatedResponse.PageSize)
	s.Equal(255, resp.PaginatedResponse.TotalRecords)
}

func (s *ClientTestSuite) TestClient_WithApplication() {
	testCases := []struct {
		name     string
		received bonsai.Application
		expect   string
	}{
		{
			name: "both Application fields filled in",
			received: bonsai.Application{
				Name:    "withName",
				Version: "withVersion",
			},
			expect: fmt.Sprintf("%s/%s %s", "withName", "withVersion", bonsai.UserAgent),
		},
		{
			name: "application name non-empty; version empty",
			received: bonsai.Application{
				Name:    "withName",
				Version: "",
			},
			expect: fmt.Sprintf("%s %s", "withName", bonsai.UserAgent),
		},
		{
			name:     "Application fields both empty",
			received: bonsai.Application{},
			expect:   bonsai.UserAgent,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			c := bonsai.NewClient(
				bonsai.WithApplication(tc.received),
			)
			s.Equal(tc.expect, c.UserAgent())
		})
	}
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}
