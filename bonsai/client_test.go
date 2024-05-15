package bonsai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"

	"github.com/go-chi/chi/v5"
	"github.com/omc/bonsai-api-go/v2/bonsai"
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

// ClientMockTestSuite is used for all requests against data
// either live from an API endpoint, or saved as a fixture from
// those same endpoints.
type ClientVCRTestSuite struct {
	recordMode           recorder.Mode
	recorder             *recorder.Recorder
	recorderUpdateRegexp *regexp.Regexp
	fileNormalizer       *regexp.Regexp
	ClientTestSuite
}

func (s *ClientVCRTestSuite) ReadOnlyRun() bool {
	return s.recordMode == recorder.ModeReplayOnly
}
func (s *ClientVCRTestSuite) WillRecord() bool {
	return !s.ReadOnlyRun()
}

func (s *ClientVCRTestSuite) SetRecorderMode(mode string) {
	switch mode {
	case "REC_ONCE":
		s.recordMode = recorder.ModeRecordOnce
	case "REC_ONLY":
		s.recordMode = recorder.ModeRecordOnly
	case "REPLAY_ONLY":
		s.recordMode = recorder.ModeReplayOnly
	case "REPLAY_WITH_NEW":
		s.recordMode = recorder.ModeReplayWithNewEpisodes
	case "PASS_THROUGH":
		s.recordMode = recorder.ModePassthrough
	default:
		s.recordMode = recorder.ModeReplayOnly
	}
}

func (s *ClientVCRTestSuite) SetupSuite() {
	var err error

	s.fileNormalizer = regexp.MustCompile("[^A-Za-z0-9-]+")

	envVCRRecordMode := os.Getenv("BONSAI_REC_MODE")
	// Passthrough if empty
	s.SetRecorderMode(envVCRRecordMode)

	if recordUpdateRegexpStr, ok := os.LookupEnv("BONSAI_REC_UPDATE_MATCHING"); ok {
		s.recorderUpdateRegexp = regexp.MustCompile(recordUpdateRegexpStr)
	}

	envKey := os.Getenv("BONSAI_API_KEY")
	envToken := os.Getenv("BONSAI_API_TOKEN")

	accessKey, err := bonsai.NewAccessKey(envKey)
	if err != nil {
		log.Fatal(fmt.Errorf("invalid user received: %w", err))
	}

	accessToken, err := bonsai.NewAccessToken(envToken)
	if err != nil {
		log.Fatal(fmt.Errorf("invalid token/password received: %w", err))
	}

	credentialPair := bonsai.CredentialPair{
		AccessKey:   accessKey,
		AccessToken: accessToken,
	}

	// If we're not passing through, and we don't have a key, panic!
	if s.WillRecord() && credentialPair.Empty() {
		log.Panic("BONSAI_API_TOKEN environment variable not set for testing")
	}

	// Configure http client and other miscellany
	s.serveMux = chi.NewRouter()
	s.client = bonsai.NewClient(
		bonsai.WithApplication(
			bonsai.Application{
				Name:    "bonsai-api-go",
				Version: "v2.3.0",
			},
		),
		bonsai.WithCredentialPair(
			credentialPair,
		),
	)

	// configure testify
	s.Assertions = require.New(s.T())
}

func (s *ClientVCRTestSuite) BeforeTest(_, testName string) {
	var err error

	s.recorder, err = recorder.NewWithOptions(
		&recorder.Options{
			// filepath is os agnostic
			CassetteName:       filepath.Join("fixtures/vcr/", s.normalize(testName)),
			Mode:               s.recordMode,
			SkipRequestLatency: true,
		},
	)

	if err != nil {
		log.Fatalf("failed to create new recorder: %+v\n", err)
	}

	if s.WillRecord() {
		// Add a hook which removes Authorization headers from all requests
		hook := func(i *cassette.Interaction) error {
			delete(i.Request.Headers, "Authorization")
			return nil
		}
		s.recorder.AddHook(hook, recorder.AfterCaptureHook)

		// Remove headers that might include secrets.
		s.recorder.AddHook(riskyHeaderFilter, recorder.AfterCaptureHook)

		s.client.SetTransport(s.recorder)
	} else {
		s.recorder.SetReplayableInteractions(true)
		s.client.SetTransport(s.recorder)
	}
}

func (s *ClientVCRTestSuite) AfterTest(_, _ string) {
	if s.recorder != nil && s.recorder.IsRecording() {
		if err := s.recorder.Stop(); err != nil {
			log.Fatalf("error stopping recorder: %v", err)
		}
	}
}

func (s *ClientVCRTestSuite) TearDownSuite() {

}

func (s *ClientVCRTestSuite) update(name string) bool {
	if s.ReadOnlyRun() {
		return false
	}

	if reflect.ValueOf(s.recorderUpdateRegexp).IsZero() {
		return true
	}

	return s.recorderUpdateRegexp.MatchString(name)
}

func (s *ClientVCRTestSuite) normalize(path string) string {
	return s.fileNormalizer.ReplaceAllLiteralString(path, "-")
}

func TestClientVCRTestSuite(t *testing.T) {
	suite.Run(t, new(ClientVCRTestSuite))
}

// ClientMockTestSuite is used for all mocked web requests.
type ClientMockTestSuite struct {
	ClientTestSuite
}

func (s *ClientMockTestSuite) SetupSuite() {
	// Configure http client and other miscellany
	s.serveMux = chi.NewRouter()
	s.server = httptest.NewServer(s.serveMux)

	s.client = bonsai.NewClient(
		bonsai.WithEndpoint(s.server.URL),
		bonsai.WithCredentialPair(
			bonsai.CredentialPair{
				AccessKey:   bonsai.AccessKey("TestKey"),
				AccessToken: bonsai.AccessToken("TestToken"),
			},
		),
	)
	// configure testify
	s.Assertions = require.New(s.T())
}

func TestClientMockTestSuite(t *testing.T) {
	suite.Run(t, new(ClientMockTestSuite))
}

func (s *ClientMockTestSuite) TestResponseErrorUnmarshallJson() {
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

func (s *ClientMockTestSuite) TestClientResponseError() {
	const p = "/clusters/doesnotexist-1234"

	// Configure Servemux to serve the error response at this path
	s.serveMux.Get(p, func(w http.ResponseWriter, _ *http.Request) {
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

func (s *ClientMockTestSuite) TestClientResponseWithPagination() {
	s.serveMux.Get("/clusters", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", bonsai.HTTPContentTypeJSON)
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

func (s *ClientMockTestSuite) TestClient_WithApplication() {
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
