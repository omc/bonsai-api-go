package bonsai

import (
	"testing"

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
}

func (s *ClientImplTestSuite) SetupSuite() {
	// configure testify
	s.Assertions = require.New(s.T())
}

func (s *ClientImplTestSuite) TestClientDefaultRateLimit() {
	c := NewClient()
	s.Equal(c.rateLimiter.Burst(), DefaultClientBurstAllowance)
	s.Equal(c.rateLimiter.Limit(), rate.Every(DefaultClientBurstDuration))
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

func TestClientImplTestSuite(t *testing.T) {
	suite.Run(t, new(ClientImplTestSuite))
}
