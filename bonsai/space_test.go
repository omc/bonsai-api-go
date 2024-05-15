package bonsai_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/omc/bonsai-api-go/v2/bonsai"
)

// Mocked Tests

func (s *ClientMockTestSuite) TestSpaceClient_All() {
	s.serveMux.Get(bonsai.SpaceAPIBasePath, func(w http.ResponseWriter, _ *http.Request) {
		respStr := `
			{
			  "spaces": [
				{
				  "path": "omc/bonsai/us-east-1/common",
				  "private_network": false,
				  "cloud": {
					"provider": "aws",
					"region": "aws-us-east-1"
				  }
				},
				{
				  "path": "omc/bonsai/eu-west-1/common",
				  "private_network": false,
				  "cloud": {
					"provider": "aws",
					"region": "aws-eu-west-1"
				  }
				},
				{
				  "path": "omc/bonsai/ap-southeast-2/common",
				  "private_network": false,
				  "cloud": {
					"provider": "aws",
					"region": "aws-ap-southeast-2"
				  }
				}
			  ]
			}
		`

		resp := &bonsai.SpacesResultList{Spaces: make([]bonsai.Space, 0, 1)}
		err := json.Unmarshal([]byte(respStr), resp)
		s.NoError(err, "successfully unmarshals json into bonsai.SpacesResultList")

		err = json.NewEncoder(w).Encode(resp)
		s.NoError(err, "successfully encodes bonsai.SpacesResultList into json")
	})

	ctx := context.Background()

	spaces, err := s.client.Space.All(ctx)
	s.NoError(err, "successfully get all spaces")
	s.Len(spaces, 3)
}

func (s *ClientMockTestSuite) TestSpaceClient_GetByPath() {
	const targetSpacePath = "omc/bonsai/us-east-1/common"

	urlPath, err := url.JoinPath(bonsai.SpaceAPIBasePath, targetSpacePath)
	s.NoError(err, "successfully create url path")

	s.serveMux.Get(urlPath, func(w http.ResponseWriter, _ *http.Request) {
		respStr := fmt.Sprintf(`
			{
			    "path": "%s",
			    "private_network": false,
			    "cloud": {
			  		"provider": "aws",
			  		"region": "aws-us-east-1"
			    }
			}
		`, targetSpacePath)

		resp := &bonsai.Space{}
		err = json.Unmarshal([]byte(respStr), resp)
		s.NoError(err, "successfully unmarshals json into bonsai.Space")

		err = json.NewEncoder(w).Encode(resp)
		s.NoError(err, "successfully encodes bonsai.Space into json")
	})

	ctx := context.Background()

	space, err := s.client.Space.GetByPath(ctx, "omc/bonsai/us-east-1/common")
	s.NoError(err, "successfully get space by path")

	s.Equal(targetSpacePath, space.Path)
	s.Equal(new(bool), space.PrivateNetwork)
	s.Equal("aws", space.Cloud.Provider)
	s.Equal("aws-us-east-1", space.Cloud.Region)
}

// VCR Tests.
func (s *ClientVCRTestSuite) TestSpaceClient_All() {
	ctx := context.Background()

	spaces, err := s.client.Space.All(ctx)
	s.NoError(err, "successfully get all spaces")
	assertGolden(s, spaces)
}

func (s *ClientVCRTestSuite) TestSpaceClient_GetByPath() {
	ctx := context.Background()

	space, err := s.client.Space.GetByPath(ctx, "omc/bonsai/eu-west-1/common")
	s.NoError(err, "successfully get space")
	assertGolden(s, space)
}
