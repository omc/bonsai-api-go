package bonsai

import (
	"encoding/json"

	"github.com/google/go-cmp/cmp"
)

func (s *ClientImplTestSuite) TestPlanAllResponseJsonUnmarshal() {
	testCases := []struct {
		name     string
		received string
		expect   planAllResponse
	}{
		{
			name: "plan example response from docs site",
			received: `
		{
			"slug": "sandbox-aws-us-east-1",
			"name": "Sandbox",
			"price_in_cents": 0,
			"billing_interval_in_months": 1,
			"single_tenant": false,
			"private_network": false,
			"available_releases": [
				"elasticsearch-7.2.0"
			],
			"available_spaces": [
				"omc/bonsai-gcp/us-east4/common",
				"omc/bonsai/ap-northeast-1/common",
				"omc/bonsai/ap-southeast-2/common",
				"omc/bonsai/eu-central-1/common",
				"omc/bonsai/eu-west-1/common",
				"omc/bonsai/us-east-1/common",
				"omc/bonsai/us-west-2/common"
			]
		}
		`,
			expect: planAllResponse{
				Slug:                    "sandbox-aws-us-east-1",
				Name:                    "Sandbox",
				PriceInCents:            0,
				BillingIntervalInMonths: 1,
				SingleTenant:            false,
				PrivateNetwork:          false,
				AvailableReleases: []string{
					"elasticsearch-7.2.0",
				},
				AvailableSpaces: []string{
					"omc/bonsai-gcp/us-east4/common",
					"omc/bonsai/ap-northeast-1/common",
					"omc/bonsai/ap-southeast-2/common",
					"omc/bonsai/eu-central-1/common",
					"omc/bonsai/eu-west-1/common",
					"omc/bonsai/us-east-1/common",
					"omc/bonsai/us-west-2/common",
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result := planAllResponse{}
			err := json.Unmarshal([]byte(tc.received), &result)
			s.NoError(err)
			s.Equal(tc.expect, result)
			s.Empty(cmp.Diff(result, tc.expect))
		})
	}
}
