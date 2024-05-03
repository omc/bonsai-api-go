package bonsai

import "net/url"

func (s *ClientImplTestSuite) TestClusterListOptsValues() {
	testCases := []struct {
		name     string
		received clusterListOpts
		expect   url.Values
	}{
		{
			name: "with populated values",
			received: clusterListOpts{
				listOpts: listOpts{
					Page: 3,
					Size: 100,
				},
				ClusterAllOpts: ClusterAllOpts{
					Query:    "a query string",
					Tenancy:  "parent",
					Location: "omc/bonsai/us-east-1/common",
				},
			},
			expect: url.Values{
				"page":     []string{"3"},
				"size":     []string{"100"},
				"q":        []string{"a query string"},
				"tenancy":  []string{"parent"},
				"location": []string{"omc/bonsai/us-east-1/common"},
			},
		},
		{
			name: "with pagination, but empty ClusterAllOpts values",
			received: clusterListOpts{
				listOpts: listOpts{
					Page: 3,
					Size: 100,
				},
			},
			expect: url.Values{
				"page": []string{"3"},
				"size": []string{"100"},
			},
		},
		{
			name:     "with empty values",
			received: clusterListOpts{},
			expect:   url.Values{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			receivedVal, err := tc.received.values()
			s.NoError(err, "received values values()")
			s.Equal(receivedVal, tc.expect)
		})
	}
}
