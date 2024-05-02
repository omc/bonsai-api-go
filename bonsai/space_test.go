package bonsai_test

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/omc/bonsai-api-go/v1/bonsai"
)

func (s *ClientTestSuite) TestSpaceClient_All() {
	s.serveMux.HandleFunc(bonsai.SpaceAPIBasePath, func(w http.ResponseWriter, _ *http.Request) {
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
