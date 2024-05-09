package bonsai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/omc/bonsai-api-go/v2/bonsai"
)

func (s *ClientMockTestSuite) TestReleaseClient_All() {
	s.serveMux.Get(bonsai.ReleaseAPIBasePath, func(w http.ResponseWriter, _ *http.Request) {
		respStr := `
		{
			"releases": [
				{
					"name": "Elasticsearch 5.6.16",
					"slug": "elasticsearch-5.6.16",
					"service_type": "elasticsearch",
					"version": "5.6.16",
					"multitenant": true
				},
				{
					"name": "Elasticsearch 6.5.4",
					"slug": "elasticsearch-6.5.4",
					"service_type": "elasticsearch",
					"version": "6.5.4",
					"multitenant": true
				},
				{
					"name": "Elasticsearch 7.2.0",
					"slug": "elasticsearch-7.2.0",
					"service_type": "elasticsearch",
					"version": "7.2.0",
					"multitenant": true
				}
			]
		}
		`

		resp := &bonsai.ReleasesResultList{Releases: make([]bonsai.Release, 0, 3)}
		err := json.Unmarshal([]byte(respStr), resp)
		s.NoError(err, "unmarshal json into bonsai.ReleasesResultList")

		err = json.NewEncoder(w).Encode(resp)
		s.NoError(err, "encode bonsai.ReleasesResultList into json")
	})

	expect := []bonsai.Release{
		{
			Name:        "Elasticsearch 5.6.16",
			Slug:        "elasticsearch-5.6.16",
			ServiceType: "elasticsearch",
			Version:     "5.6.16",
			MultiTenant: true,
		},
		{
			Name:        "Elasticsearch 6.5.4",
			Slug:        "elasticsearch-6.5.4",
			ServiceType: "elasticsearch",
			Version:     "6.5.4",
			MultiTenant: true,
		},
		{
			Name:        "Elasticsearch 7.2.0",
			Slug:        "elasticsearch-7.2.0",
			ServiceType: "elasticsearch",
			Version:     "7.2.0",
			MultiTenant: true,
		},
	}
	releases, err := s.client.Release.All(context.Background())
	s.NoError(err, "successfully get all releases")
	s.Len(releases, 3)

	s.ElementsMatch(expect, releases, "elements in expect match elements in received releases")
}

func (s *ClientMockTestSuite) TestReleaseClient_GetBySlug() {
	const targetReleaseSlug = "elasticsearch-7.2.0"

	urlPath, err := url.JoinPath(bonsai.ReleaseAPIBasePath, targetReleaseSlug)
	s.NoError(err, "successfully resolved path")

	s.serveMux.Get(urlPath, func(w http.ResponseWriter, _ *http.Request) {
		respStr := fmt.Sprintf(`
		{
			"name": "Elasticsearch 7.2.0",
			"slug": "%s",
			"service_type": "elasticsearch",
			"version": "7.2.0",
			"multitenant": true
		}
		`, targetReleaseSlug)

		resp := &bonsai.Release{}
		err = json.Unmarshal([]byte(respStr), resp)
		s.NoError(err, "unmarshals json into bonsai.Release")

		err = json.NewEncoder(w).Encode(resp)
		s.NoError(err, "encodes bonsai.Release into json on the writer")
	})

	expect := bonsai.Release{
		Slug:        "elasticsearch-7.2.0",
		Name:        "Elasticsearch 7.2.0",
		ServiceType: "elasticsearch",
		Version:     "7.2.0",
		MultiTenant: true,
	}

	resultResp, err := s.client.Release.GetBySlug(context.Background(), targetReleaseSlug)
	s.NoError(err, "successfully get release by path")

	s.Equal(expect, resultResp, "elements in expect match elements in received release response")
}

// VCR Tests.
func (s *ClientVCRTestSuite) TestReleaseClient_All() {
	ctx := context.Background()

	spaces, err := s.client.Release.All(ctx)
	s.NoError(err, "successfully get all spaces")
	assertGolden(s, spaces)
}

func (s *ClientVCRTestSuite) TestReleaseClient_GetByPath() {
	ctx := context.Background()

	space, err := s.client.Release.GetBySlug(ctx, "opensearch-2.6.0-mt")
	s.NoError(err, "successfully get space")
	assertGolden(s, space)
}
