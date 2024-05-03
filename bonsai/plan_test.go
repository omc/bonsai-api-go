package bonsai_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/omc/bonsai-api-go/v1/bonsai"
)

func (s *ClientTestSuite) TestPlanClient_All() {
	s.serveMux.HandleFunc(bonsai.PlanAPIBasePath, func(w http.ResponseWriter, _ *http.Request) {
		respStr := `
		{
		"plans": [
			{
				  "slug": "sandbox-aws-us-east-1",
				  "name": "Sandbox",
				  "price_in_cents": 0,
				  "billing_interval_in_months": 1,
				  "single_tenant": false,
				  "private_network": false,
				  "available_releases": [
					  "7.2.0"
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
			  },
			  {
				 "slug": "standard-sm",
				 "name": "Standard Small",
				 "price_in_cents": 5000,
				 "billing_interval_in_months": 1,
				 "single_tenant": false,
				 "private_network": false,
				 "available_releases": [
					"elasticsearch-5.6.16",
					"elasticsearch-6.8.3",
					"elasticsearch-7.2.0"
				 ],
				 "available_spaces": [
					"omc/bonsai/ap-northeast-1/common",
					"omc/bonsai/ap-southeast-2/common",
					"omc/bonsai/eu-central-1/common",
					"omc/bonsai/eu-west-1/common",
					"omc/bonsai/us-east-1/common",
					"omc/bonsai/us-west-2/common"
				 ]
			  }
			]
		}	
		`
		_, err := w.Write([]byte(respStr))
		s.NoError(err, "write respStr to http.ResponseWriter")
	})

	expect := []bonsai.Plan{
		{
			Slug:                    "sandbox-aws-us-east-1",
			Name:                    "Sandbox",
			PriceInCents:            0,
			BillingIntervalInMonths: 1,
			SingleTenant:            false,
			PrivateNetwork:          false,
			AvailableReleases: []bonsai.Release{
				// TODO: we'll see whether the response is actually a
				// shortened version like this or a slug
				// the documentation is conflicting at
				// https://bonsai.io/docs/plans-api-introduction
				{Slug: "7.2.0"},
			},
			AvailableSpaces: []bonsai.Space{
				{Path: "omc/bonsai-gcp/us-east4/common"},
				{Path: "omc/bonsai/ap-northeast-1/common"},
				{Path: "omc/bonsai/ap-southeast-2/common"},
				{Path: "omc/bonsai/eu-central-1/common"},
				{Path: "omc/bonsai/eu-west-1/common"},
				{Path: "omc/bonsai/us-east-1/common"},
				{Path: "omc/bonsai/us-west-2/common"},
			},
		},
		{
			Slug:                    "standard-sm",
			Name:                    "Standard Small",
			PriceInCents:            5000,
			BillingIntervalInMonths: 1,
			SingleTenant:            false,
			PrivateNetwork:          false,
			AvailableReleases: []bonsai.Release{
				{Slug: "elasticsearch-5.6.16"},
				{Slug: "elasticsearch-6.8.3"},
				{Slug: "elasticsearch-7.2.0"},
			},
			AvailableSpaces: []bonsai.Space{
				{Path: "omc/bonsai/ap-northeast-1/common"},
				{Path: "omc/bonsai/ap-southeast-2/common"},
				{Path: "omc/bonsai/eu-central-1/common"},
				{Path: "omc/bonsai/eu-west-1/common"},
				{Path: "omc/bonsai/us-east-1/common"},
				{Path: "omc/bonsai/us-west-2/common"},
			},
		},
	}
	spaces, err := s.client.Plan.All(context.Background())
	s.NoError(err, "successfully get all spaces")
	s.Len(spaces, 2)

	s.ElementsMatch(expect, spaces, "elements expected match elements in received spaces")
}

func (s *ClientTestSuite) TestPlanClient_GetByPath() {
	const targetPlanPath = "sandbox-aws-us-east-1"

	urlPath, err := url.JoinPath(bonsai.PlanAPIBasePath, "sandbox-aws-us-east-1")
	s.NoError(err, "successfully resolved path")

	respStr := fmt.Sprintf(`
		{
			"slug": "%s",
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
		`, targetPlanPath)

	s.serveMux.HandleFunc(urlPath, func(w http.ResponseWriter, _ *http.Request) {
		_, err = w.Write([]byte(respStr))
		s.NoError(err, "wrote response string to writer")
	})

	expect := bonsai.Plan{
		Slug:                    "sandbox-aws-us-east-1",
		Name:                    "Sandbox",
		PriceInCents:            0,
		BillingIntervalInMonths: 1,
		SingleTenant:            false,
		PrivateNetwork:          false,
		AvailableReleases: []bonsai.Release{
			{Slug: "elasticsearch-7.2.0"},
		},
		AvailableSpaces: []bonsai.Space{
			{Path: "omc/bonsai-gcp/us-east4/common"},
			{Path: "omc/bonsai/ap-northeast-1/common"},
			{Path: "omc/bonsai/ap-southeast-2/common"},
			{Path: "omc/bonsai/eu-central-1/common"},
			{Path: "omc/bonsai/eu-west-1/common"},
			{Path: "omc/bonsai/us-east-1/common"},
			{Path: "omc/bonsai/us-west-2/common"},
		},
	}

	resultResp, err := s.client.Plan.GetBySlug(context.Background(), "sandbox-aws-us-east-1")
	s.NoError(err, "successfully get space by path")

	s.Equal(expect, resultResp, "expected struct matches unmarshaled result")
}
